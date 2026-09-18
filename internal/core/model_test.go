package core

import (
	"testing"
	"time"
)

func TestParseTenureDays(t *testing.T) {
	if ParseTenureDays("1 năm 2 tháng 3 ngày") <= ParseTenureDays("11 tháng 29 ngày") {
		t.Fatal("tenure must sort by duration")
	}
}

func TestManualShiftWins(t *testing.T) {
	b := DefaultBusinessSettings()
	b.ManualShifts[ManualShiftKey("u1", "Pick")] = "Ca 2"
	rows := []PayrollRow{{
		Job: "Pick", EvenOdd: "Chẵn", DO: "1", User: "u1", Name: "A",
		Start: time.Date(2026, 9, 1, 8, 0, 0, 0, time.Local),
		End: time.Date(2026, 9, 1, 9, 0, 0, 0, time.Local),
		Pieces: 100, Duration: 1,
	}}
	_, _, s := BuildTables(rows, b, time.Now())
	if len(s.Rows) != 1 || s.Rows[0][6] != "Ca 2" {
		t.Fatalf("manual shift not applied: %#v", s.Rows)
	}
}

func TestEvenCreditNoDeduct(t *testing.T) {
	c := &catAgg{DO: map[string]bool{"a": true, "b": true}, SKU: 100}
	if evenCredit(c, false) != 2 {
		t.Fatal("one credit per DO expected")
	}
}

func TestPercentType(t *testing.T) {
	if FormatPercent(0.42) != "42.0%" {
		t.Fatal(FormatPercent(0.42))
	}
}
