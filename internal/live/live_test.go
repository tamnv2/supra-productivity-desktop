package live

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func syntheticXLSX(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	files := map[string]string{
		"xl/worksheets/sheet1.xml": `<?xml version="1.0" encoding="UTF-8"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>
<row r="1">
<c r="A1" t="inlineStr"><is><t>Loại công việc</t></is></c>
<c r="B1" t="inlineStr"><is><t>Chẵn/Lẻ</t></is></c>
<c r="C1" t="inlineStr"><is><t>Mã DO</t></is></c>
<c r="D1" t="inlineStr"><is><t>Nhân viên</t></is></c>
<c r="E1" t="inlineStr"><is><t>Họ và Tên</t></is></c>
<c r="F1" t="inlineStr"><is><t>Đối tác</t></is></c>
<c r="G1" t="inlineStr"><is><t>SiteId</t></is></c>
<c r="H1" t="inlineStr"><is><t>Thời gian bắt đầu</t></is></c>
<c r="I1" t="inlineStr"><is><t>Thời gian kết thúc</t></is></c>
<c r="J1" t="inlineStr"><is><t>Thời gian thực hiện (Phút)</t></is></c>
<c r="K1" t="inlineStr"><is><t>Tổng SKU đã thực hiện</t></is></c>
<c r="L1" t="inlineStr"><is><t>Tổng sản lượng đã thực hiện (Pieces)</t></is></c>
</row>
<row r="2">
<c r="A2" t="inlineStr"><is><t>Pick</t></is></c>
<c r="B2" t="inlineStr"><is><t>Chẵn</t></is></c>
<c r="C2" t="inlineStr"><is><t>DO-1</t></is></c>
<c r="D2" t="inlineStr"><is><t>picker01</t></is></c>
<c r="E2" t="inlineStr"><is><t>Test User</t></is></c>
<c r="F2" t="inlineStr"><is><t>Inhouse</t></is></c>
<c r="G2"><v>1291</v></c>
<c r="H2" t="inlineStr"><is><t>2026-09-19 08:00:00</t></is></c>
<c r="I2" t="inlineStr"><is><t>2026-09-19 09:00:00</t></is></c>
<c r="J2"><v>60</v></c>
<c r="K2"><v>20</v></c>
<c r="L2"><v>300</v></c>
</row>
</sheetData></worksheet>`,
	}
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil { t.Fatal(err) }
		if _, err = io.WriteString(w, body); err != nil { t.Fatal(err) }
	}
	if err := zw.Close(); err != nil { t.Fatal(err) }
	return buf.Bytes()
}

func TestParsePayrollXLSX(t *testing.T) {
	rows, err := ParsePayrollXLSX(syntheticXLSX(t), Binding{})
	if err != nil { t.Fatal(err) }
	if len(rows) != 1 { t.Fatalf("rows=%d", len(rows)) }
	r := rows[0]
	if r.User != "picker01" || r.Pieces != 300 || r.SKU != 20 || r.Site != "1291" {
		t.Fatalf("unexpected row: %#v", r)
	}
	if r.Duration != 60 {
		t.Fatalf("V1.3 duration must remain payroll minutes, got %v", r.Duration)
	}
}

func TestParseActiveJSONWithFieldMap(t *testing.T) {
	data := []byte(`{"payload":{"rows":[{"state":"running","u":"picker01","done":25,"total":50,"progress":50}]}}`)
	b := Binding{
		DataPath: "payload.rows",
		FieldMap: map[string]string{
			"status": "state",
			"user": "u",
			"sku_done": "done",
			"sku_total": "total",
			"sku_progress": "progress",
		},
	}
	table, err := ParseActiveJSON(data, b)
	if err != nil { t.Fatal(err) }
	if len(table.Rows) != 1 { t.Fatalf("rows=%d", len(table.Rows)) }
	if table.Rows[0][6] != "picker01" { t.Fatalf("user=%v", table.Rows[0][6]) }
	if got, ok := table.Rows[0][15].(float64); !ok || got != 0.5 {
		t.Fatalf("progress=%#v", table.Rows[0][15])
	}
}

func TestExecuteAppliesSessionWithoutLeakingIntoBinding(t *testing.T) {
	var gotAuth, gotToken string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotToken = r.Header.Get("token")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	data, meta, err := Execute(ctx, srv.Client(), Binding{URL: srv.URL}, Session{Authorization: "Bearer synthetic", Token: "synthetic-token"})
	if err != nil { t.Fatal(err) }
	if string(data) != `{"ok":true}` || meta.StatusCode != 200 { t.Fatalf("bad response: %s %#v", data, meta) }
	if gotAuth != "Bearer synthetic" || gotToken != "synthetic-token" { t.Fatalf("session headers not applied") }
}

func TestBuildPeopleTableDeduplicatesUsers(t *testing.T) {
	payroll, err := ParsePayrollXLSX(syntheticXLSX(t), Binding{})
	if err != nil { t.Fatal(err) }
	payroll = append(payroll, payroll[0])
	table := BuildPeopleTable(payroll)
	if len(table.Rows) != 1 { t.Fatalf("rows=%d", len(table.Rows)) }
}


func TestExpandBindingRelativeDates(t *testing.T) {
	now := time.Date(2026, 9, 19, 8, 30, 0, 0, time.Local)
	b := Binding{
		URL: "https://example.invalid/report?from={{YESTERDAY_ISO}}&to={{TODAY_ISO}}",
		Body: "next={{TOMORROW_DMY}}",
		Headers: map[string]string{"X-Date": "{{TODAY_DMY}}"},
	}
	got := ExpandBinding(b, now)
	if got.URL != "https://example.invalid/report?from=2026-09-18&to=2026-09-19" {
		t.Fatalf("url=%q", got.URL)
	}
	if got.Body != "next=20/09/2026" {
		t.Fatalf("body=%q", got.Body)
	}
	if got.Headers["X-Date"] != "19/09/2026" {
		t.Fatalf("header=%q", got.Headers["X-Date"])
	}
}


func TestEnrichPayrollAndUserPDAFromActive(t *testing.T) {
	payroll, err := ParsePayrollXLSX(syntheticXLSX(t), Binding{})
	if err != nil { t.Fatal(err) }
	payroll[0].Name = ""
	payroll[0].MNV = ""
	activeJSON := []byte(`{"data":[{"UserName":"picker01","FullName":"Active Name","EmployeeId":"E001","Provider":"Inhouse","SiteId":"1291","Tenure":"2 tháng","DeviceId":"PDA-07","Client":"HY1","Status":"running"}]}`)
	active, err := ParseActiveJSON(activeJSON, Binding{})
	if err != nil { t.Fatal(err) }
	enriched := EnrichPayrollFromActive(payroll, active)
	if enriched[0].Name != "Active Name" || enriched[0].MNV != "E001" {
		t.Fatalf("enrichment failed: %#v", enriched[0])
	}
	people := BuildPeopleTableCombined(enriched, active)
	if len(people.Rows) != 1 { t.Fatalf("people rows=%d", len(people.Rows)) }
	if people.Rows[0][2] != "picker01" || people.Rows[0][6] != "PDA-07" {
		t.Fatalf("people row=%#v", people.Rows[0])
	}
}
