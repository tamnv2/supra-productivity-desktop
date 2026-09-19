package core

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

// BusinessSettings mirrors the recovered V1.3 stable executable behavior.
// Values here are operational defaults recovered from the stable binary; secrets
// and private endpoint bindings are intentionally kept outside the public repo.
type BusinessSettings struct {
	ActiveStatus          string            `json:"active_status"`
	PickShift             string            `json:"pick_shift"`
	PackShift             string            `json:"pack_shift"`
	PickEvenQuota         int               `json:"pick_even_quota"`
	PackEvenQuota         int               `json:"pack_even_quota"`
	PickDeductSKU         bool              `json:"pick_deduct_sku"`
	PackDeductSKU         bool              `json:"pack_deduct_sku"`
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
	Skip20                map[string]bool   `json:"skip_20_minutes"`
	ManualShifts          map[string]string `json:"manual_shifts"`
}

func DefaultBusinessSettings() BusinessSettings {
	return BusinessSettings{
		ActiveStatus:          "Tất cả",
		PickShift:             "Ca 2",
		PackShift:             "Tất cả",
		PickEvenQuota:         8,
		PackEvenQuota:         5,
		PickDeductSKU:         true,
		PackDeductSKU:         true,
		PickRequireEven:       true,
		PickTargetInhouseEven: 8550,
		PickTargetInhouseOdd:  2250,
		PickTargetOtherEven:   10000,
		PickTargetOtherOdd:    2100,
		PickSpeedInhouseEven:  17,
		PickSpeedInhouseOdd:   7,
		PickSpeedOtherEven:    17,
		PickSpeedOtherOdd:     7,
		CheckStart:            "07:00",
		CheckEnd:              "12:48",
		CheckLimit:            150,
		Skip20:                map[string]bool{},
		ManualShifts:          map[string]string{},
	}
}

func NormalizeBusinessSettings(b *BusinessSettings) {
	d := DefaultBusinessSettings()
	if b.ActiveStatus == "" {
		b.ActiveStatus = d.ActiveStatus
	}
	if b.PickShift == "" {
		b.PickShift = d.PickShift
	}
	if b.PackShift == "" {
		b.PackShift = d.PackShift
	}
	if b.PickEvenQuota < 1 {
		b.PickEvenQuota = d.PickEvenQuota
	}
	if b.PackEvenQuota < 1 {
		b.PackEvenQuota = d.PackEvenQuota
	}
	if b.PickTargetInhouseEven < 1 {
		b.PickTargetInhouseEven = d.PickTargetInhouseEven
	}
	if b.PickTargetInhouseOdd < 1 {
		b.PickTargetInhouseOdd = d.PickTargetInhouseOdd
	}
	if b.PickTargetOtherEven < 1 {
		b.PickTargetOtherEven = d.PickTargetOtherEven
	}
	if b.PickTargetOtherOdd < 1 {
		b.PickTargetOtherOdd = d.PickTargetOtherOdd
	}
	if b.PickSpeedInhouseEven < 1 {
		b.PickSpeedInhouseEven = d.PickSpeedInhouseEven
	}
	if b.PickSpeedInhouseOdd < 1 {
		b.PickSpeedInhouseOdd = d.PickSpeedInhouseOdd
	}
	if b.PickSpeedOtherEven < 1 {
		b.PickSpeedOtherEven = d.PickSpeedOtherEven
	}
	if b.PickSpeedOtherOdd < 1 {
		b.PickSpeedOtherOdd = d.PickSpeedOtherOdd
	}
	if b.CheckStart == "" {
		b.CheckStart = d.CheckStart
	}
	if b.CheckEnd == "" {
		b.CheckEnd = d.CheckEnd
	}
	if b.CheckLimit < 1 {
		b.CheckLimit = d.CheckLimit
	}
	if b.Skip20 == nil {
		b.Skip20 = map[string]bool{}
	}
	if b.ManualShifts == nil {
		b.ManualShifts = map[string]string{}
	}
}

type PayrollRow struct {
	Job, EvenOdd, DO, User, Name, Provider, MNV, Site, Tenure string
	Shift, ManualShift                                       string
	Start, End                                               time.Time
	Duration                                                 float64 // minutes, matching V1.3 payroll export
	SKU, Pieces                                              int
}

type Table struct {
	Headers []string
	Rows    [][]any
}

type aggKey struct{ User, Job string }

type catAgg struct {
	DO       map[string]bool
	Pieces   int
	Duration float64
	SKUByDO  map[string]int
	Last     time.Time
}

type userAgg struct {
	User, Job, Name, Shift, Manual string
	First, Last                    time.Time
	Cats                           map[string]*catAgg
}

type employeeInfo struct {
	Name, MNV, Provider, Site, Tenure, Shift, ManualShift string
}

func ManualShiftKey(user, job string) string {
	return strings.ToLower(strings.TrimSpace(user)) + "|" + strings.ToLower(strings.TrimSpace(job))
}

func validShift(s string) bool {
	return s == "Ca 1" || s == "Ca 2" || s == "Ca HC"
}

func normalizeText(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	r := strings.NewReplacer(
		"à", "a", "á", "a", "ạ", "a", "ả", "a", "ã", "a", "â", "a", "ầ", "a", "ấ", "a", "ậ", "a", "ẩ", "a", "ẫ", "a",
		"ă", "a", "ằ", "a", "ắ", "a", "ặ", "a", "ẳ", "a", "ẵ", "a",
		"è", "e", "é", "e", "ẹ", "e", "ẻ", "e", "ẽ", "e", "ê", "e", "ề", "e", "ế", "e", "ệ", "e", "ể", "e", "ễ", "e",
		"ì", "i", "í", "i", "ị", "i", "ỉ", "i", "ĩ", "i",
		"ò", "o", "ó", "o", "ọ", "o", "ỏ", "o", "õ", "o", "ô", "o", "ồ", "o", "ố", "o", "ộ", "o", "ổ", "o", "ỗ", "o",
		"ơ", "o", "ờ", "o", "ớ", "o", "ợ", "o", "ở", "o", "ỡ", "o",
		"ù", "u", "ú", "u", "ụ", "u", "ủ", "u", "ũ", "u", "ư", "u", "ừ", "u", "ứ", "u", "ự", "u", "ử", "u", "ữ", "u",
		"ỳ", "y", "ý", "y", "ỵ", "y", "ỷ", "y", "ỹ", "y", "đ", "d",
	)
	return strings.Join(strings.Fields(r.Replace(s)), " ")
}

func isEven(s string) bool {
	n := normalizeText(s)
	return n == "chan" || n == "even" || strings.Contains(n, "chan")
}

func isOdd(s string) bool {
	n := normalizeText(s)
	return n == "le" || n == "odd" || strings.Contains(n, "le")
}

func isPickJob(s string) bool {
	n := normalizeText(s)
	return strings.Contains(n, "lay hang") || strings.Contains(n, "pick")
}

func isPackJob(s string) bool {
	n := normalizeText(s)
	return strings.Contains(n, "dong goi") || strings.Contains(n, "pack")
}

func isInhouse(provider string) bool {
	return strings.EqualFold(strings.TrimSpace(provider), "Inhouse")
}

func autoShift(job, provider string, start, end, now time.Time) string {
	p := strings.ToLower(strings.TrimSpace(provider))
	if strings.EqualFold(provider, "robotics") || strings.Contains(p, "robotics") {
		return "Ca 1"
	}
	minute := start.Hour()*60 + start.Minute()
	if !start.IsZero() && !end.IsZero() && minute >= 450 && minute <= 510 && end.Hour() >= 20 {
		return "Ca HC"
	}
	if isPackJob(job) {
		if minute < 540 {
			return "Ca 1"
		}
		if minute < 780 && !end.IsZero() && end.Hour() > 19 {
			return "Ca 2"
		}
		if minute < 780 {
			return "Ca 1"
		}
		return "Ca 2"
	}
	if minute < 480 {
		return "Ca 1"
	}
	if minute > 779 {
		return "Ca 2"
	}
	if !end.IsZero() && end.Hour() > 19 {
		return "Ca 2"
	}
	if now.Hour() < 14 {
		return "Ca 1"
	}
	return "Ca 2"
}

func shiftClass(shift string, when time.Time) string {
	minute := when.Hour()*60 + when.Minute()
	switch shift {
	case "Ca 1":
		if minute < 840 {
			return "in"
		}
		return "ot"
	case "Ca 2":
		if minute >= 840 {
			return "in"
		}
		return "ot"
	case "Ca HC":
		if minute >= 450 && minute < 1050 {
			return "in"
		}
		return "ot"
	default:
		return "in"
	}
}

func catKey(evenOdd, class string) string {
	switch {
	case isEven(evenOdd):
		return "chan|" + class
	case isOdd(evenOdd):
		return "le|" + class
	default:
		return normalizeText(evenOdd) + "|" + class
	}
}

func cat(a *userAgg, evenOdd, class string) *catAgg {
	if a == nil {
		return &catAgg{DO: map[string]bool{}, SKUByDO: map[string]int{}}
	}
	key := catKey(evenOdd, class)
	if c := a.Cats[key]; c != nil {
		return c
	}
	c := &catAgg{DO: map[string]bool{}, SKUByDO: map[string]int{}}
	a.Cats[key] = c
	return c
}

func firstNonEmpty(v ...string) string {
	for _, s := range v {
		if strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}

func safeRatio(a, b float64) float64 {
	if b <= 0 {
		return 0
	}
	return a / b
}

func relativeEnd(now, end time.Time) string {
	if end.IsZero() {
		return ""
	}
	minutes := int(now.Sub(end).Minutes())
	if minutes < 0 {
		minutes = 0
	}
	if minutes > 59 {
		return fmt.Sprintf("%d giờ %d phút trước · tính từ %s", minutes/60, minutes%60, end.Format("15:04"))
	}
	return fmt.Sprintf("%d phút trước · tính từ %s", minutes, end.Format("15:04"))
}

func countDO(c *catAgg) int {
	if c == nil {
		return 0
	}
	return len(c.DO)
}

func speed(c *catAgg) int {
	if c == nil || c.Duration <= 0 {
		return 0
	}
	return int(float64(c.Pieces) / c.Duration)
}

func BuildTables(rows []PayrollRow, b BusinessSettings, now time.Time) (Table, Table, Table) {
	NormalizeBusinessSettings(&b)

	refs := map[string]employeeInfo{}
	for _, r := range rows {
		if strings.TrimSpace(r.User) == "" {
			continue
		}
		old := refs[r.User]
		refs[r.User] = employeeInfo{
			Name:        firstNonEmpty(r.Name, old.Name),
			MNV:         firstNonEmpty(r.MNV, old.MNV),
			Provider:    firstNonEmpty(r.Provider, old.Provider),
			Site:        firstNonEmpty(r.Site, old.Site),
			Tenure:      firstNonEmpty(r.Tenure, old.Tenure),
			Shift:       firstNonEmpty(r.Shift, old.Shift),
			ManualShift: firstNonEmpty(r.ManualShift, old.ManualShift),
		}
	}

	aggs := map[aggKey]*userAgg{}
	for _, r := range rows {
		if strings.TrimSpace(r.User) == "" {
			continue
		}
		key := aggKey{User: r.User, Job: r.Job}
		a := aggs[key]
		if a == nil {
			a = &userAgg{User: r.User, Job: r.Job, Name: r.Name, Cats: map[string]*catAgg{}}
			aggs[key] = a
		}
		if a.Name == "" {
			a.Name = r.Name
		}
		if a.First.IsZero() || (!r.Start.IsZero() && r.Start.Before(a.First)) {
			a.First = r.Start
		}
		if r.End.After(a.Last) {
			a.Last = r.End
		}
	}

	for _, a := range aggs {
		ref := refs[a.User]
		manual := b.ManualShifts[ManualShiftKey(a.User, a.Job)]
		if manual == "" || manual == "Tự động" {
			manual = ref.ManualShift
		}
		a.Manual = manual
		if validShift(manual) {
			a.Shift = manual
		} else if validShift(ref.Shift) {
			a.Shift = ref.Shift
		} else {
			a.Shift = autoShift(a.Job, ref.Provider, a.First, a.Last, now)
		}
	}

	for _, r := range rows {
		a := aggs[aggKey{User: r.User, Job: r.Job}]
		if a == nil {
			continue
		}
		class := shiftClass(a.Shift, r.End)
		c := cat(a, r.EvenOdd, class)
		doCode := strings.TrimSpace(r.DO)
		if doCode == "" {
			doCode = fmt.Sprintf("%s-%d", r.User, len(c.DO)+1)
		}
		c.DO[doCode] = true
		c.Pieces += r.Pieces
		c.Duration += r.Duration
		c.SKUByDO[doCode] += r.SKU
		if r.End.After(c.Last) {
			c.Last = r.End
		}
	}

	return buildPickTable(aggs, refs, b, now), buildPackTable(aggs, refs, b), buildShiftTable(aggs, refs, b, now)
}

func sortedAggKeys(aggs map[aggKey]*userAgg, predicate func(string) bool) []aggKey {
	keys := make([]aggKey, 0, len(aggs))
	for k, a := range aggs {
		if predicate(a.Job) {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].User == keys[j].User {
			return keys[i].Job < keys[j].Job
		}
		return keys[i].User < keys[j].User
	})
	return keys
}

func buildPickTable(aggs map[aggKey]*userAgg, refs map[string]employeeInfo, b BusinessSettings, now time.Time) Table {
	headers := []string{
		"Họ và tên", "Tuổi nghề", "Tỉ lệ sản lượng trong ca", "Mã nhân viên", "User",
		"Nhà cung cấp", "Site", "Phân ca", "SL chẵn cần xử lí", "DO chẵn", "SL chẵn",
		"DO lẻ", "SL lẻ ", "DO chẵn tăng ca", "SL chẵn tăng ca", "DO lẻ tăng ca",
		"SL lẻ tăng ca", "Thời gian kết thúc đơn cuối cùng", "Tốc độ pick chẵn trong ca",
		"Tốc độ pick lẻ trong ca", "Tốc độ pick chẵn tăng ca", "Tốc độ pick lẻ tăng ca",
		" SL lẻ còn thiếu so với target", "Bỏ qua kiểm tra 20 phút", "DL_PICK20_FLAG", "", "KT 1C1L",
	}
	t := Table{Headers: headers}

	for _, key := range sortedAggKeys(aggs, isPickJob) {
		a := aggs[key]
		if b.PickShift != "Tất cả" && a.Shift != b.PickShift {
			continue
		}
		ref := refs[a.User]
		if !b.ShowAllSite && strings.TrimSpace(ref.Site) != "" && strings.TrimSpace(ref.Site) != "1921" {
			continue
		}

		eIn, oIn := cat(a, "chẵn", "in"), cat(a, "lẻ", "in")
		eOT, oOT := cat(a, "chẵn", "ot"), cat(a, "lẻ", "ot")
		targetEven, targetOdd := b.PickTargetOtherEven, b.PickTargetOtherOdd
		if isInhouse(ref.Provider) {
			targetEven, targetOdd = b.PickTargetInhouseEven, b.PickTargetInhouseOdd
		}

		progress := safeRatio(float64(eIn.Pieces), float64(targetEven)) + safeRatio(float64(oIn.Pieces), float64(targetOdd))
		evenCredit := eIn.Pieces
		if b.PickDeductSKU {
			for _, sku := range eIn.SKUByDO {
				evenCredit -= int(math.Ceil(float64(sku) / 10))
				if evenCredit < 0 {
					evenCredit = 0
				}
			}
		}
		missingOdd := targetOdd - oIn.Pieces
		if missingOdd < 0 {
			missingOdd = 0
		}

		last := a.Last
		for _, c := range []*catAgg{eIn, oIn, eOT, oOT} {
			if c.Last.After(last) {
				last = c.Last
			}
		}

		skipKey := ManualShiftKey(a.User, a.Job)
		skipped := b.Skip20[skipKey]
		flag := ""
		if skipped {
			flag = "OK"
		}
		check1C1L := "OK"
		if b.Enable1C1L && len(eIn.DO) > 0 && b.PickEvenQuota > 0 && eIn.Pieces/b.PickEvenQuota < len(eIn.DO) {
			check1C1L = "Lỗi 1C1L"
		}

		t.Rows = append(t.Rows, []any{
			firstNonEmpty(ref.Name, a.Name), ref.Tenure, progress, ref.MNV, a.User, ref.Provider,
			siteAny(ref.Site), a.Shift, evenCredit, countDO(eIn), eIn.Pieces, countDO(oIn), oIn.Pieces,
			countDO(eOT), eOT.Pieces, countDO(oOT), oOT.Pieces, relativeEnd(now, last),
			speed(eIn), speed(oIn), speed(eOT), speed(oOT), missingOdd, skipped, flag, "", check1C1L,
		})
	}
	return t
}

func buildPackTable(aggs map[aggKey]*userAgg, refs map[string]employeeInfo, b BusinessSettings) Table {
	t := Table{Headers: []string{"Họ và tên", "User", "Site", "Phân ca", "DO chẵn", "SL chẵn", "NSLD chẵn", "DO lẻ", "SL lẻ", "NSLD lẻ"}}
	for _, key := range sortedAggKeys(aggs, isPackJob) {
		a := aggs[key]
		if b.PackShift != "Tất cả" && a.Shift != b.PackShift {
			continue
		}
		ref := refs[a.User]
		even, odd := cat(a, "chẵn", "in"), cat(a, "lẻ", "in")
		t.Rows = append(t.Rows, []any{
			firstNonEmpty(ref.Name, a.Name), a.User, siteAny(ref.Site), a.Shift,
			countDO(even), even.Pieces, speed(even), countDO(odd), odd.Pieces, speed(odd),
		})
	}
	return t
}

func buildShiftTable(aggs map[aggKey]*userAgg, refs map[string]employeeInfo, b BusinessSettings, now time.Time) Table {
	t := Table{Headers: []string{
		"Họ và tên", "Mã nhân viên", "User", "Nhà cung cấp", "Loại công việc",
		"Phân ca tự động", "Phân ca thủ công", "Thời gian bắt đầu đơn đầu tiên",
		"Thời gian kết thúc đơn cuối cùng", "Ghi chú phân ca",
	}}
	for _, key := range sortedAggKeys(aggs, func(string) bool { return true }) {
		a := aggs[key]
		ref := refs[a.User]
		auto := autoShift(a.Job, ref.Provider, a.First, a.Last, now)
		note := ""
		if validShift(a.Manual) {
			note = "Ưu tiên phân ca thủ công"
		}
		t.Rows = append(t.Rows, []any{
			firstNonEmpty(ref.Name, a.Name), ref.MNV, a.User, ref.Provider, a.Job,
			auto, a.Manual, a.First, a.Last, note,
		})
	}
	return t
}

func siteAny(raw string) any {
	if n, err := strconv.Atoi(strings.TrimSpace(raw)); err == nil {
		return n
	}
	return raw
}

func FormatPercent(v any) string {
	f, ok := ToFloat(v)
	if !ok {
		return fmt.Sprint(v)
	}
	return fmt.Sprintf("%.1f%%", f*100)
}

func ToFloat(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case string:
		f, e := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(x, "%")), 64)
		if e != nil {
			return 0, false
		}
		if strings.Contains(x, "%") {
			f /= 100
		}
		return f, true
	}
	return 0, false
}

// ParseTenureDays keeps typed tenure sorting used by the desktop list views.
func ParseTenureDays(s string) int {
	s = strings.ToLower(strings.TrimSpace(s))
	total := 0
	fields := strings.Fields(s)
	for i := 0; i < len(fields)-1; i++ {
		n, e := strconv.Atoi(fields[i])
		if e != nil {
			continue
		}
		u := fields[i+1]
		switch {
		case strings.HasPrefix(u, "năm") || strings.HasPrefix(u, "nam") || strings.HasPrefix(u, "year"):
			total += n * 365
		case strings.HasPrefix(u, "tháng") || strings.HasPrefix(u, "thang") || strings.HasPrefix(u, "month"):
			total += n * 30
		case strings.HasPrefix(u, "ngày") || strings.HasPrefix(u, "ngay") || strings.HasPrefix(u, "day"):
			total += n
		}
	}
	return total
}
