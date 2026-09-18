package core

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

// BusinessSettings contains only non-secret operational preferences.
// Production values are persisted locally on the workstation, never in the public repository.
type BusinessSettings struct {
	ActiveStatus          string            `json:"active_status"`
	PickShift             string            `json:"pick_shift"`
	PackShift             string            `json:"pack_shift"`
	PickEvenQuota         int               `json:"pick_even_quota"`
	PickDeductSKU         bool              `json:"pick_deduct_sku"`
	PickRequireEven       bool              `json:"pick_require_even"`
	PickTargetInhouseEven int               `json:"pick_target_inhouse_even"`
	PickTargetInhouseOdd  int               `json:"pick_target_inhouse_odd"`
	PickTargetOtherEven   int               `json:"pick_target_other_even"`
	PickTargetOtherOdd    int               `json:"pick_target_other_odd"`
	PickSpeedInhouseEven  int               `json:"pick_speed_inhouse_even"`
	PickSpeedInhouseOdd   int               `json:"pick_speed_inhouse_odd"`
	PickSpeedOtherEven    int               `json:"pick_speed_other_even"`
	PickSpeedOtherOdd     int               `json:"pick_speed_other_odd"`
	Enable1C1L            bool              `json:"enable_1c1l"`
	ShowIncompleteEven    bool              `json:"show_incomplete_even_only"`
	Show1C1LErrors        bool              `json:"show_1c1l_errors_only"`
	CheckStart            string            `json:"check_1c1l_start"`
	CheckEnd              string            `json:"check_1c1l_end"`
	CheckLimit            int               `json:"check_1c1l_limit"`
	ShowAllSite           bool              `json:"show_all_site"`
	PrimarySite           string            `json:"primary_site"`
	Skip20                map[string]bool   `json:"skip_20_minutes"`
	ManualShifts          map[string]string `json:"manual_shifts"`
}

func DefaultBusinessSettings() BusinessSettings {
	return BusinessSettings{
		ActiveStatus: "Tất cả", PickShift: "Tất cả", PackShift: "Tất cả",
		PickEvenQuota: 0, PickDeductSKU: true, PickRequireEven: true,
		PickTargetInhouseEven: 1, PickTargetInhouseOdd: 1,
		PickTargetOtherEven: 1, PickTargetOtherOdd: 1,
		PickSpeedInhouseEven: 1, PickSpeedInhouseOdd: 1,
		PickSpeedOtherEven: 1, PickSpeedOtherOdd: 1,
		CheckStart: "07:00", CheckEnd: "13:00", CheckLimit: 0,
		Skip20: map[string]bool{}, ManualShifts: map[string]string{},
	}
}

func NormalizeBusinessSettings(b *BusinessSettings) {
	d := DefaultBusinessSettings()
	if b.ActiveStatus == "" { b.ActiveStatus = d.ActiveStatus }
	if b.PickShift == "" { b.PickShift = d.PickShift }
	if b.PackShift == "" { b.PackShift = d.PackShift }
	if b.PickTargetInhouseEven <= 0 { b.PickTargetInhouseEven = d.PickTargetInhouseEven }
	if b.PickTargetInhouseOdd <= 0 { b.PickTargetInhouseOdd = d.PickTargetInhouseOdd }
	if b.PickTargetOtherEven <= 0 { b.PickTargetOtherEven = d.PickTargetOtherEven }
	if b.PickTargetOtherOdd <= 0 { b.PickTargetOtherOdd = d.PickTargetOtherOdd }
	if b.PickSpeedInhouseEven <= 0 { b.PickSpeedInhouseEven = d.PickSpeedInhouseEven }
	if b.PickSpeedInhouseOdd <= 0 { b.PickSpeedInhouseOdd = d.PickSpeedInhouseOdd }
	if b.PickSpeedOtherEven <= 0 { b.PickSpeedOtherEven = d.PickSpeedOtherEven }
	if b.PickSpeedOtherOdd <= 0 { b.PickSpeedOtherOdd = d.PickSpeedOtherOdd }
	if b.CheckStart == "" { b.CheckStart = d.CheckStart }
	if b.CheckEnd == "" { b.CheckEnd = d.CheckEnd }
	if b.Skip20 == nil { b.Skip20 = map[string]bool{} }
	if b.ManualShifts == nil { b.ManualShifts = map[string]string{} }
}

type PayrollRow struct {
	Job, EvenOdd, DO, User, Name, Provider, MNV, Site, Tenure string
	Start, End time.Time
	Duration float64
	SKU, Pieces int
}

type Table struct { Headers []string; Rows [][]any }
type aggKey struct { User, Job string }
type catAgg struct { DO map[string]bool; Pieces, SKU int; Duration float64; Last time.Time }
type userAgg struct { User, Job, Name string; First, Last time.Time; Shift, Manual string; Cats map[string]*catAgg }

func catKey(evenOdd, bucket string) string { return strings.ToLower(strings.TrimSpace(evenOdd)) + "|" + bucket }
func isEven(s string) bool { n:=strings.ToLower(strings.TrimSpace(s)); return strings.Contains(n,"chẵn") || strings.Contains(n,"chan") || n=="even" }
func isOdd(s string) bool { n:=strings.ToLower(strings.TrimSpace(s)); return strings.Contains(n,"lẻ") || strings.Contains(n,"le") || n=="odd" }
func isPickJob(s string) bool { n:=strings.ToLower(strings.TrimSpace(s)); return strings.Contains(n,"pick") || strings.Contains(n,"lấy") || strings.Contains(n,"lay") }
func isPackJob(s string) bool { n:=strings.ToLower(strings.TrimSpace(s)); return strings.Contains(n,"pack") || strings.Contains(n,"đóng") || strings.Contains(n,"dong") }
func validShift(s string) bool { return s=="Ca 1" || s=="Ca 2" || s=="Ca HC" }

func ManualShiftKey(user, job string) string {
	return strings.ToLower(strings.TrimSpace(user)) + "|" + strings.ToLower(strings.TrimSpace(job))
}

func autoShift(job,user string, first,last,timeNow time.Time) string {
	if strings.EqualFold(strings.TrimSpace(user),"robotics") { return "Ca 1" }
	if !first.IsZero() && !last.IsZero() {
		fmins:=first.Hour()*60+first.Minute()
		lmins:=last.Hour()*60+last.Minute()
		if fmins>=7*60+30 && fmins<=8*60+30 && lmins>=15*60+30 && lmins<=17*60+30 { return "Ca HC" }
		if fmins < 14*60 { return "Ca 1" }
		return "Ca 2"
	}
	if timeNow.Hour() < 14 { return "Ca 1" }
	return "Ca 2"
}

func bucketFor(r PayrollRow, shift string) string {
	if r.Start.IsZero() { return "in" }
	switch shift {
	case "Ca 1":
		if r.Start.Hour() >= 15 { return "ot" }
	case "Ca 2":
		if r.Start.Hour() < 14 { return "ot" }
	case "Ca HC":
		if r.Start.Hour() >= 18 { return "ot" }
	}
	return "in"
}

func BuildTables(rows []PayrollRow, b BusinessSettings, now time.Time) (Table,Table,Table) {
	NormalizeBusinessSettings(&b)
	groups:=map[aggKey]*userAgg{}
	meta:=map[string]PayrollRow{}
	for _,r:=range rows {
		if strings.TrimSpace(r.User)=="" { continue }
		k:=aggKey{r.User,r.Job}; g:=groups[k]
		if g==nil { g=&userAgg{User:r.User,Job:r.Job,Name:r.Name,Cats:map[string]*catAgg{}}; groups[k]=g }
		if g.Name=="" { g.Name=r.Name }
		if g.First.IsZero() || (!r.Start.IsZero() && r.Start.Before(g.First)) { g.First=r.Start }
		if r.End.After(g.Last) { g.Last=r.End }
		meta[r.User]=r
	}
	for _,g:=range groups {
		if m:=b.ManualShifts[ManualShiftKey(g.User,g.Job)]; validShift(m) {
			g.Manual=m; g.Shift=m
		} else {
			g.Shift=autoShift(g.Job,g.User,g.First,g.Last,now)
		}
	}
	for _,r:=range rows {
		g:=groups[aggKey{r.User,r.Job}]; if g==nil { continue }
		bucket:=bucketFor(r,g.Shift)
		eo:="other"
		if isEven(r.EvenOdd){eo="even"} else if isOdd(r.EvenOdd){eo="odd"}
		ck:=catKey(eo,bucket); c:=g.Cats[ck]
		if c==nil { c=&catAgg{DO:map[string]bool{}}; g.Cats[ck]=c }
		if r.DO!="" { c.DO[r.DO]=true }
		c.Pieces+=r.Pieces; c.SKU+=r.SKU; c.Duration+=r.Duration
		if r.End.After(c.Last){c.Last=r.End}
	}
	return buildPick(groups,meta,b,now), buildPack(groups,meta,b), buildShift(groups,meta)
}

func getCat(g *userAgg, eo,bucket string)*catAgg{
	c:=g.Cats[catKey(eo,bucket)]
	if c==nil{return &catAgg{DO:map[string]bool{}}}
	return c
}
func doCount(c *catAgg)int{return len(c.DO)}
func speed(c *catAgg)int{if c.Duration<=0{return 0}; return int(math.Round(float64(c.Pieces)/c.Duration))}
func evenCredit(c *catAgg,deduct bool)int{
	if !deduct{return doCount(c)}
	if doCount(c)==0{return 0}
	return maxInt(doCount(c),int(math.Ceil(float64(c.SKU)/20.0)))
}
func maxInt(a,b int)int{if a>b{return a};return b}
func targetFor(provider string,b BusinessSettings)(int,int,int,int){
	if strings.EqualFold(strings.TrimSpace(provider),"Inhouse"){
		return b.PickTargetInhouseEven,b.PickTargetInhouseOdd,b.PickSpeedInhouseEven,b.PickSpeedInhouseOdd
	}
	return b.PickTargetOtherEven,b.PickTargetOtherOdd,b.PickSpeedOtherEven,b.PickSpeedOtherOdd
}
func check1C1L(even,odd int,skip bool,b BusinessSettings)string{
	if !b.Enable1C1L{return "Tắt"}
	if skip{return "Bỏ qua 20p"}
	if even>0&&odd>0{return "OK"}
	return "Thiếu"
}
func relativeEnd(t,now time.Time)string{
	if t.IsZero(){return ""}
	d:=now.Sub(t)
	if d<0{return t.Format("15:04:05")}
	if d<time.Hour{return fmt.Sprintf("%d phút trước",int(d.Minutes()))}
	return t.Format("15:04:05")
}

func buildPick(groups map[aggKey]*userAgg,meta map[string]PayrollRow,b BusinessSettings,now time.Time)Table{
	h:=[]string{"Họ và tên","Tuổi nghề","Tỉ lệ sản lượng trong ca","Mã nhân viên","User","Nhà cung cấp","Site","Phân ca","SL chẵn cần xử lí","DO chẵn","SL chẵn","DO lẻ","SL lẻ","DO chẵn tăng ca","SL chẵn tăng ca","DO lẻ tăng ca","SL lẻ tăng ca","Thời gian kết thúc đơn cuối cùng","Tốc độ pick chẵn trong ca","Tốc độ pick lẻ trong ca","Tốc độ pick chẵn tăng ca","Tốc độ pick lẻ tăng ca","SL lẻ còn thiếu so với target","Bỏ qua kiểm tra 20 phút","KT 1C1L"}
	t:=Table{Headers:h}
	keys:=make([]aggKey,0,len(groups))
	for k,g:=range groups{if isPickJob(g.Job){keys=append(keys,k)}}
	sort.Slice(keys,func(i,j int)bool{return strings.ToLower(groups[keys[i]].Name)<strings.ToLower(groups[keys[j]].Name)})
	for _,k:=range keys {
		g:=groups[k]
		if b.PickShift!="Tất cả"&&g.Shift!=b.PickShift{continue}
		m:=meta[g.User]
		if !b.ShowAllSite&&b.PrimarySite!=""&&m.Site!=""&&m.Site!=b.PrimarySite{continue}
		e:=getCat(g,"even","in");o:=getCat(g,"odd","in");eot:=getCat(g,"even","ot");oot:=getCat(g,"odd","ot")
		te,to,_,_:=targetFor(m.Provider,b)
		ratio:=0.0
		if te>0{ratio+=float64(e.Pieces)/float64(te)}
		if to>0{ratio+=float64(o.Pieces)/float64(to)}
		need:=b.PickEvenQuota-evenCredit(e,b.PickDeductSKU); if need<0{need=0}
		var needVal any=""
		if b.PickRequireEven{if need==0{needVal="OK"}else{needVal=need}}
		skip:=b.Skip20[g.User]
		status:=check1C1L(doCount(e),doCount(o),skip,b)
		if b.ShowIncompleteEven&&need==0{continue}
		if b.Show1C1LErrors&&status=="OK"{continue}
		remaining:=0
		if doCount(o)>0&&te>0&&to>0{
			remaining=int(math.Ceil(float64(to)*(1-float64(e.Pieces)/float64(te))-float64(o.Pieces)))
			if remaining<0{remaining=0}
		}
		last:=g.Last
		row:=[]any{firstNonEmpty(g.Name,m.Name),m.Tenure,ratio,m.MNV,g.User,m.Provider,m.Site,g.Shift,needVal,doCount(e),e.Pieces,doCount(o),o.Pieces,doCount(eot),eot.Pieces,doCount(oot),oot.Pieces,relativeEnd(last,now),speed(e),speed(o),speed(eot),speed(oot),remaining,skip,status}
		t.Rows=append(t.Rows,row)
	}
	return t
}

func buildPack(groups map[aggKey]*userAgg,meta map[string]PayrollRow,b BusinessSettings)Table{
	h:=[]string{"Họ và tên","User","Site","Phân ca","DO chẵn","SL chẵn","NSLD chẵn","DO lẻ","SL lẻ","NSLD lẻ"}
	t:=Table{Headers:h};keys:=[]aggKey{}
	for k,g:=range groups{if isPackJob(g.Job){keys=append(keys,k)}}
	sort.Slice(keys,func(i,j int)bool{return strings.ToLower(groups[keys[i]].Name)<strings.ToLower(groups[keys[j]].Name)})
	for _,k:=range keys{
		g:=groups[k]
		if b.PackShift!="Tất cả"&&g.Shift!=b.PackShift{continue}
		m:=meta[g.User];e:=getCat(g,"even","in");o:=getCat(g,"odd","in")
		t.Rows=append(t.Rows,[]any{firstNonEmpty(g.Name,m.Name),g.User,m.Site,g.Shift,doCount(e),e.Pieces,speed(e),doCount(o),o.Pieces,speed(o)})
	}
	return t
}

func buildShift(groups map[aggKey]*userAgg,meta map[string]PayrollRow)Table{
	h:=[]string{"Họ và tên","Mã nhân viên","User","Nhà cung cấp","Loại công việc","Phân ca tự động","Phân ca thủ công","Thời gian bắt đầu đơn","Thời gian kết thúc đơn","Ghi chú phân ca"}
	t:=Table{Headers:h};keys:=make([]aggKey,0,len(groups))
	for k:=range groups{keys=append(keys,k)}
	sort.Slice(keys,func(i,j int)bool{return groups[keys[i]].First.Before(groups[keys[j]].First)})
	for _,k:=range keys{
		g:=groups[k];m:=meta[g.User]
		auto:=autoShift(g.Job,g.User,g.First,g.Last,time.Now())
		note:="";if g.Manual!=""{note="Ưu tiên phân ca thủ công"}
		t.Rows=append(t.Rows,[]any{firstNonEmpty(g.Name,m.Name),m.MNV,g.User,m.Provider,g.Job,auto,g.Manual,g.First,g.Last,note})
	}
	return t
}
func firstNonEmpty(v ...string)string{for _,s:=range v{if strings.TrimSpace(s)!=""{return s}};return ""}

func FormatPercent(v any) string {
	f,ok:=ToFloat(v)
	if !ok{return fmt.Sprint(v)}
	return fmt.Sprintf("%.1f%%",f*100)
}
func ToFloat(v any)(float64,bool){
	switch x:=v.(type){
	case float64:return x,true
	case float32:return float64(x),true
	case int:return float64(x),true
	case int64:return float64(x),true
	case string:
		f,e:=strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(x,"%")),64)
		if e!=nil{return 0,false}
		if strings.Contains(x,"%"){f/=100}
		return f,true
	}
	return 0,false
}

// ParseTenureDays makes tenure sorting chronological rather than lexical.
func ParseTenureDays(s string) int {
	s=strings.ToLower(strings.TrimSpace(s)); total:=0
	fields:=strings.Fields(s)
	for i:=0;i<len(fields)-1;i++{
		n,e:=strconv.Atoi(fields[i]);if e!=nil{continue}
		u:=fields[i+1]
		switch{
		case strings.HasPrefix(u,"năm")||strings.HasPrefix(u,"nam")||strings.HasPrefix(u,"year"): total+=n*365
		case strings.HasPrefix(u,"tháng")||strings.HasPrefix(u,"thang")||strings.HasPrefix(u,"month"): total+=n*30
		case strings.HasPrefix(u,"ngày")||strings.HasPrefix(u,"ngay")||strings.HasPrefix(u,"day"): total+=n
		}
	}
	return total
}
