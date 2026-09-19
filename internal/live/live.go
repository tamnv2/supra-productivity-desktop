package live

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tamnv2/supra-productivity-desktop/internal/core"
)

type Binding struct {
	Method       string            `json:"method"`
	URL          string            `json:"url"`
	Body         string            `json:"body,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
	ResponseKind string            `json:"response_kind,omitempty"`
	DataPath     string            `json:"data_path,omitempty"`
	Sheet        string            `json:"sheet,omitempty"`
	HeaderRow    int               `json:"header_row,omitempty"`
	FieldMap     map[string]string `json:"field_map,omitempty"`
}

type Session struct {
	Authorization string
	Token         string
	APISID        string
	USID          string
	Signature     string
	Nonce         string
	UserAgent     string
	Headers       map[string]string
}

type HTTPMeta struct {
	StatusCode int
	Bytes      int
	Elapsed    time.Duration
}

func ExpandBinding(b Binding, now time.Time) Binding {
	repl := map[string]string{
		"{{TODAY_ISO}}": now.Format("2006-01-02"),
		"{{TOMORROW_ISO}}": now.AddDate(0, 0, 1).Format("2006-01-02"),
		"{{TODAY_DMY}}": now.Format("02/01/2006"),
		"{{TOMORROW_DMY}}": now.AddDate(0, 0, 1).Format("02/01/2006"),
		"{{NOW_ISO}}": now.Format(time.RFC3339),
		"{{BUSINESS_DATE}}": now.Format("2006-01-02"),
	}
	expand := func(s string) string {
		for k, v := range repl {
			s = strings.ReplaceAll(s, k, v)
		}
		return s
	}
	b.URL = expand(b.URL)
	b.Body = expand(b.Body)
	if len(b.Headers) > 0 {
		h := make(map[string]string, len(b.Headers))
		for k, v := range b.Headers { h[k] = expand(v) }
		b.Headers = h
	}
	return b
}

func Execute(ctx context.Context, client *http.Client, b Binding, s Session) ([]byte, HTTPMeta, error) {
	if client == nil {
		client = &http.Client{Timeout: 45 * time.Second}
	}
	method := strings.ToUpper(strings.TrimSpace(b.Method))
	if method == "" {
		method = http.MethodGet
	}
	if _, err := url.ParseRequestURI(strings.TrimSpace(b.URL)); err != nil {
		return nil, HTTPMeta{}, fmt.Errorf("invalid binding URL: %w", err)
	}
	var body io.Reader
	if b.Body != "" {
		body = strings.NewReader(b.Body)
	}
	req, err := http.NewRequestWithContext(ctx, method, b.URL, body)
	if err != nil {
		return nil, HTTPMeta{}, err
	}
	for k, v := range b.Headers {
		if strings.TrimSpace(k) != "" {
			req.Header.Set(k, v)
		}
	}
	applySession(req, s)
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return nil, HTTPMeta{Elapsed: time.Since(start)}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 100*1024*1024))
	meta := HTTPMeta{StatusCode: resp.StatusCode, Bytes: len(data), Elapsed: time.Since(start)}
	if err != nil {
		return nil, meta, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, meta, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return data, meta, nil
}

func applySession(req *http.Request, s Session) {
	if s.Authorization != "" {
		req.Header.Set("Authorization", s.Authorization)
	}
	if s.Token != "" && req.Header.Get("token") == "" {
		req.Header.Set("token", s.Token)
	}
	if s.APISID != "" && req.Header.Get("APISID") == "" {
		req.Header.Set("APISID", s.APISID)
	}
	if s.USID != "" && req.Header.Get("USID") == "" {
		req.Header.Set("USID", s.USID)
	}
	if s.Signature != "" {
		req.Header.Set("x-signature", s.Signature)
	}
	if s.Nonce != "" {
		req.Header.Set("x-signature-nonce", s.Nonce)
	}
	if s.UserAgent != "" {
		req.Header.Set("User-Agent", s.UserAgent)
	}
	for k, v := range s.Headers {
		if strings.TrimSpace(k) == "" || isSensitiveHeader(k) {
			continue
		}
		if req.Header.Get(k) == "" {
			req.Header.Set(k, v)
		}
	}
}

func isSensitiveHeader(k string) bool {
	n := strings.ToLower(strings.TrimSpace(k))
	return n == "authorization" || n == "cookie" || n == "token" ||
		n == "apisid" || n == "usid" || strings.Contains(n, "signature") ||
		strings.Contains(n, "password")
}

func ParsePayrollXLSX(data []byte, b Binding) ([]core.PayrollRow, error) {
	rows, err := readXLSXRows(data, b.Sheet)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("XLSX has no rows")
	}
	headerRow := b.HeaderRow
	if headerRow <= 0 {
		headerRow = detectPayrollHeader(rows)
	}
	if headerRow <= 0 || headerRow > len(rows) {
		return nil, fmt.Errorf("cannot detect payroll header")
	}
	headers := rows[headerRow-1]
	index := mapHeaders(headers)
	resolve := func(canonical string, aliases ...string) int {
		if b.FieldMap != nil {
			if v := strings.TrimSpace(b.FieldMap[canonical]); v != "" {
				if i, ok := index[normalize(v)]; ok {
					return i
				}
			}
		}
		for _, a := range aliases {
			if i, ok := index[normalize(a)]; ok {
				return i
			}
		}
		return -1
	}
	idx := struct {
		job, evenOdd, doCode, user, name, provider, mnv, site, tenure int
		start, end, duration, sku, pieces int
	}{
		resolve("job", "Loại công việc", "Job", "Work Type"),
		resolve("even_odd", "Chẵn/Lẻ", "Chan/Le", "Even/Odd"),
		resolve("do_code", "Mã DO", "DO", "DO Code"),
		resolve("user", "Nhân viên", "User", "Employee User"),
		resolve("name", "Họ và Tên", "Họ tên", "Name", "Employee Name"),
		resolve("provider", "Đối tác", "Nhà cung cấp", "Provider", "Vendor"),
		resolve("mnv", "Mã nhân viên", "MNV", "Employee ID"),
		resolve("site", "SiteId", "Site", "Site ID"),
		resolve("tenure", "Tuổi nghề", "Tenure"),
		resolve("start", "Thời gian bắt đầu", "Start Time", "Start"),
		resolve("end", "Thời gian kết thúc", "End Time", "End"),
		resolve("duration", "Thời gian thực hiện (Phút)", "Thời gian thực hiện", "Duration Minutes", "Duration"),
		resolve("sku", "Tổng SKU đã thực hiện", "SKU đã lấy", "SKU Done", "SKU"),
		resolve("pieces", "Tổng sản lượng đã thực hiện (Pieces)", "SL đã lấy", "Pieces Done", "Pieces"),
	}
	if idx.job < 0 || idx.user < 0 || idx.start < 0 {
		return nil, fmt.Errorf("payroll header missing required columns")
	}
	var out []core.PayrollRow
	for _, row := range rows[headerRow:] {
		get := func(i int) string {
			if i < 0 || i >= len(row) {
				return ""
			}
			return strings.TrimSpace(row[i])
		}
		user := get(idx.user)
		if user == "" {
			continue
		}
		start, _ := parseDate(get(idx.start))
		end, _ := parseDate(get(idx.end))
		minutes, _ := parseNumber(get(idx.duration))
		sku, _ := parseInt(get(idx.sku))
		pieces, _ := parseInt(get(idx.pieces))
		out = append(out, core.PayrollRow{
			Job: get(idx.job), EvenOdd: get(idx.evenOdd), DO: get(idx.doCode), User: user,
			Name: get(idx.name), Provider: get(idx.provider), MNV: get(idx.mnv),
			Site: get(idx.site), Tenure: get(idx.tenure), Start: start, End: end,
			Duration: minutes / 60.0, SKU: sku, Pieces: pieces,
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("payroll XLSX parsed zero data rows")
	}
	return out, nil
}

func ParseActiveJSON(data []byte, b Binding) (core.Table, error) {
	var root any
	if err := json.Unmarshal(data, &root); err != nil {
		return core.Table{}, err
	}
	node := root
	if p := strings.TrimSpace(b.DataPath); p != "" {
		var ok bool
		node, ok = lookupPath(root, p)
		if !ok {
			return core.Table{}, fmt.Errorf("active data_path not found")
		}
	}
	items := findObjectSlice(node)
	if len(items) == 0 {
		return core.Table{}, fmt.Errorf("active JSON has no object array")
	}
	type field struct {
		Key string
		Header string
		Aliases []string
	}
	fields := []field{
		{"status","Trạng thái",[]string{"status","state","trangthai"}},
		{"alert","Cảnh báo",[]string{"alert","warning","canhbao"}},
		{"order_type","Loại đơn",[]string{"ordertype","order_type","loaidon"}},
		{"store_code","Mã cửa hàng",[]string{"storecode","store_code","macuahang"}},
		{"do_code","Mã DO",[]string{"docode","do_code","ma_do","do"}},
		{"user","User",[]string{"user","username","picker","employee"}},
		{"name","Họ và tên",[]string{"name","fullname","employee_name","hoten"}},
		{"mnv","Mã nhân viên",[]string{"mnv","employeeid","employee_id"}},
		{"provider","Nhà cung cấp",[]string{"provider","vendor","nhacungcap"}},
		{"site","Site",[]string{"site","siteid","site_id"}},
		{"shift","Phân ca",[]string{"shift","ca","phanca"}},
		{"tenure","Tuổi nghề",[]string{"tenure","tuoinghe"}},
		{"sku_done","SKU đã lấy",[]string{"skudone","sku_done","pickedsku"}},
		{"sku_total","Tổng SKU",[]string{"skutotal","sku_total","totalsku"}},
		{"sku_progress","Tiến độ SKU",[]string{"skuprogress","sku_progress"}},
		{"qty_done","SL đã lấy",[]string{"qtydone","qty_done","piecesdone","pickedqty"}},
		{"qty_total","Tổng SL",[]string{"qtytotal","qty_total","piecestotal","totalqty"}},
		{"qty_progress","Tiến độ SL",[]string{"qtyprogress","qty_progress","piecesprogress"}},
		{"elapsed","Thời gian đã dùng",[]string{"elapsed","elapsedtime","timeused"}},
		{"expected","Thời gian kỳ vọng",[]string{"expected","expectedtime"}},
		{"delta","Chênh lệch",[]string{"delta","difference","timedelta"}},
		{"device","Thiết bị",[]string{"device","pda","deviceid"}},
		{"client","Client",[]string{"client","clientid"}},
		{"start_date","Ngày đầu làm việc",[]string{"startdate","start_date","firstworkdate"}},
	}
	headers := make([]string, len(fields))
	for i, f := range fields { headers[i] = f.Header }
	t := core.Table{Headers: headers}
	for _, item := range items {
		row := make([]any, len(fields))
		for i, f := range fields {
			spec := ""
			if b.FieldMap != nil { spec = b.FieldMap[f.Key] }
			v := valueBySpec(item, spec, f.Aliases)
			switch f.Key {
			case "sku_done","sku_total","qty_done","qty_total":
				if n, ok := numberAny(v); ok { row[i] = int(n) } else { row[i] = stringify(v) }
			case "sku_progress","qty_progress":
				if n, ok := numberAny(v); ok {
					if n > 1 { n /= 100 }
					row[i] = n
				} else { row[i] = parsePercentAny(v) }
			default:
				row[i] = stringify(v)
			}
		}
		t.Rows = append(t.Rows, row)
	}
	return t, nil
}

func BuildPeopleTable(payroll []core.PayrollRow) core.Table {
	headers := []string{"Họ và tên","Mã nhân viên","User","Nhà cung cấp","Site","Tuổi nghề"}
	byUser := map[string]core.PayrollRow{}
	for _, r := range payroll {
		if strings.TrimSpace(r.User) == "" { continue }
		prev, ok := byUser[r.User]
		if !ok || (prev.MNV == "" && r.MNV != "") || (prev.Name == "" && r.Name != "") {
			byUser[r.User] = r
		}
	}
	users := make([]string,0,len(byUser))
	for u := range byUser { users=append(users,u) }
	sort.Slice(users,func(i,j int)bool{return strings.ToLower(users[i])<strings.ToLower(users[j])})
	t:=core.Table{Headers:headers}
	for _,u:=range users{
		r:=byUser[u]
		t.Rows=append(t.Rows,[]any{r.Name,r.MNV,r.User,r.Provider,r.Site,r.Tenure})
	}
	return t
}

func mapHeaders(headers []string) map[string]int {
	out:=map[string]int{}
	for i,h:=range headers {
		if n:=normalize(h); n!="" { out[n]=i }
	}
	return out
}
func detectPayrollHeader(rows [][]string) int {
	required:=[]string{normalize("Loại công việc"),normalize("Nhân viên"),normalize("Thời gian bắt đầu")}
	bestRow,bestScore:=0,0
	limit:=len(rows);if limit>30{limit=30}
	for i:=0;i<limit;i++{
		idx:=mapHeaders(rows[i]);score:=0
		for _,r:=range required{if _,ok:=idx[r];ok{score++}}
		for _,a:=range []string{"Mã DO","Chẵn/Lẻ","Tổng sản lượng đã thực hiện (Pieces)"}{if _,ok:=idx[normalize(a)];ok{score++}}
		if score>bestScore{bestScore=score;bestRow=i+1}
	}
	if bestScore<3{return 0}
	return bestRow
}

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)
func normalize(s string) string {
	repl:=strings.NewReplacer("à","a","á","a","ạ","a","ả","a","ã","a","â","a","ầ","a","ấ","a","ậ","a","ẩ","a","ẫ","a","ă","a","ằ","a","ắ","a","ặ","a","ẳ","a","ẵ","a","è","e","é","e","ẹ","e","ẻ","e","ẽ","e","ê","e","ề","e","ế","e","ệ","e","ể","e","ễ","e","ì","i","í","i","ị","i","ỉ","i","ĩ","i","ò","o","ó","o","ọ","o","ỏ","o","õ","o","ô","o","ồ","o","ố","o","ộ","o","ổ","o","ỗ","o","ơ","o","ờ","o","ớ","o","ợ","o","ở","o","ỡ","o","ù","u","ú","u","ụ","u","ủ","u","ũ","u","ư","u","ừ","u","ứ","u","ự","u","ử","u","ữ","u","ỳ","y","ý","y","ỵ","y","ỷ","y","ỹ","y","đ","d")
	return nonAlnum.ReplaceAllString(repl.Replace(strings.ToLower(strings.TrimSpace(s))),"")
}

func parseNumber(s string)(float64,bool){
	s=strings.TrimSpace(strings.ReplaceAll(s,",",""))
	if s==""{return 0,false}
	f,e:=strconv.ParseFloat(s,64);return f,e==nil
}
func parseInt(s string)(int,bool){f,ok:=parseNumber(s);return int(f),ok}
func parseDate(s string)(time.Time,bool){
	s=strings.TrimSpace(s);if s==""{return time.Time{},false}
	if f,e:=strconv.ParseFloat(s,64);e==nil && f>1000{
		base:=time.Date(1899,12,30,0,0,0,0,time.Local)
		whole,frac:=mathModf(f)
		return base.AddDate(0,0,int(whole)).Add(time.Duration(frac*24*float64(time.Hour))),true
	}
	for _,layout:=range []string{"2006-01-02 15:04:05","2006-01-02T15:04:05","02/01/2006 15:04:05","02/01/2006 15:04","02/01/2006","1/2/2006 15:04:05"}{
		if t,e:=time.ParseInLocation(layout,s,time.Local);e==nil{return t,true}
	}
	return time.Time{},false
}
func mathModf(v float64)(float64,float64){i:=float64(int64(v));return i,v-i}

func lookupPath(root any, p string)(any,bool){
	cur:=root
	for _,part:=range strings.Split(p,"."){
		switch x:=cur.(type){
		case map[string]any:
			v,ok:=x[part];if !ok{return nil,false};cur=v
		default:return nil,false
		}
	}
	return cur,true
}
func findObjectSlice(v any) []map[string]any {
	switch x:=v.(type){
	case []any:
		out:=make([]map[string]any,0,len(x))
		for _,it:=range x{if m,ok:=it.(map[string]any);ok{out=append(out,m)}}
		if len(out)>0{return out}
	case map[string]any:
		for _,k:=range []string{"data","items","rows","result","results","content","list"}{
			if child,ok:=x[k];ok{if r:=findObjectSlice(child);len(r)>0{return r}}
		}
		for _,child:=range x{if r:=findObjectSlice(child);len(r)>0{return r}}
	}
	return nil
}
func valueBySpec(item map[string]any,spec string,aliases []string) any {
	if strings.TrimSpace(spec)!=""{
		if v,ok:=lookupPath(item,spec);ok{return v}
		if v,ok:=item[spec];ok{return v}
	}
	norm:=map[string]any{}
	for k,v:=range item{norm[normalize(k)]=v}
	for _,a:=range aliases{if v,ok:=norm[normalize(a)];ok{return v}}
	return nil
}
func stringify(v any) string {
	if v==nil{return ""}
	switch x:=v.(type){
	case string:return strings.TrimSpace(x)
	case json.Number:return x.String()
	case float64:
		if x==float64(int64(x)){return strconv.FormatInt(int64(x),10)}
		return strconv.FormatFloat(x,'f',2,64)
	case bool:
		if x{return "Có"};return "Không"
	}
	return fmt.Sprint(v)
}
func numberAny(v any)(float64,bool){
	switch x:=v.(type){
	case float64:return x,true
	case float32:return float64(x),true
	case int:return float64(x),true
	case int64:return float64(x),true
	case json.Number:f,e:=x.Float64();return f,e==nil
	case string:return parseNumber(strings.TrimSuffix(strings.TrimSpace(x),"%"))
	}
	return 0,false
}
func parsePercentAny(v any) any {
	s:=stringify(v)
	if s==""{return ""}
	n,ok:=parseNumber(strings.TrimSuffix(s,"%"));if !ok{return s}
	if strings.Contains(s,"%")||n>1{n/=100}
	return n
}

type sharedStringsXML struct{ SI []struct{ T string `xml:"t"`; R []struct{T string `xml:"t"`} `xml:"r"` } `xml:"si"` }
type worksheetXML struct{ Rows []struct{ Cells []struct{ R string `xml:"r,attr"`; T string `xml:"t,attr"`; V string `xml:"v"`; IS struct{T string `xml:"t"`} `xml:"is"` } `xml:"c"` } `xml:"sheetData>row"` }
type workbookXML struct{ Sheets []struct{Name string `xml:"name,attr"`; RID string `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships id,attr"`} `xml:"sheets>sheet"` }
type relsXML struct{ Rel []struct{ID string `xml:"Id,attr"`; Target string `xml:"Target,attr"`} `xml:"Relationship"` }

func readXLSXRows(data []byte, sheetName string)([][]string,error){
	z,e:=zip.NewReader(bytes.NewReader(data),int64(len(data)));if e!=nil{return nil,e}
	files:=map[string]*zip.File{};for _,f:=range z.File{files[f.Name]=f}
	shared:=[]string{}
	if f:=files["xl/sharedStrings.xml"];f!=nil{
		var sx sharedStringsXML;if e=decodeZipXML(f,&sx);e==nil{
			for _,si:=range sx.SI{s:=si.T;for _,r:=range si.R{s+=r.T};shared=append(shared,s)}
		}
	}
	target:="xl/worksheets/sheet1.xml"
	if strings.TrimSpace(sheetName)!=""{
		if wb:=files["xl/workbook.xml"];wb!=nil{
			var w workbookXML;_ = decodeZipXML(wb,&w)
			var rel relsXML;if rf:=files["xl/_rels/workbook.xml.rels"];rf!=nil{_ = decodeZipXML(rf,&rel)}
			relmap:=map[string]string{};for _,r:=range rel.Rel{relmap[r.ID]=r.Target}
			for _,s:=range w.Sheets{
				if strings.EqualFold(strings.TrimSpace(s.Name),strings.TrimSpace(sheetName)){
					if t:=relmap[s.RID];t!=""{target=path.Clean("xl/"+strings.TrimPrefix(t,"/"))}
				}
			}
		}
	}
	f:=files[target];if f==nil{return nil,fmt.Errorf("worksheet not found: %s",target)}
	var ws worksheetXML;if e=decodeZipXML(f,&ws);e!=nil{return nil,e}
	out:=make([][]string,0,len(ws.Rows))
	for _,r:=range ws.Rows{
		max:=0;for _,c:=range r.Cells{col:=cellCol(c.R);if col>max{max=col}}
		row:=make([]string,max+1)
		for _,c:=range r.Cells{
			col:=cellCol(c.R);val:=c.V
			switch c.T{
			case "s":
				if i,er:=strconv.Atoi(c.V);er==nil&&i>=0&&i<len(shared){val=shared[i]}
			case "inlineStr":val=c.IS.T
			case "b":if c.V=="1"{val="TRUE"}else{val="FALSE"}
			}
			if col>=0&&col<len(row){row[col]=val}
		}
		out=append(out,row)
	}
	return out,nil
}
func decodeZipXML(f *zip.File,v any)error{r,e:=f.Open();if e!=nil{return e};defer r.Close();return xml.NewDecoder(r).Decode(v)}
func cellCol(ref string)int{
	n:=0;found:=false
	for _,r:=range ref{
		if r>='A'&&r<='Z'{n=n*26+int(r-'A'+1);found=true}else if r>='a'&&r<='z'{n=n*26+int(r-'a'+1);found=true}else{break}
	}
	if !found{return 0}
	return n-1
}
