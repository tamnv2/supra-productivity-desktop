package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/tamnv2/supra-productivity-desktop/internal/core"
)

type DateRange struct {
	From time.Time
	To   time.Time
}

type daySnapshot struct {
	Date      string            `json:"date"`
	UpdatedAt time.Time         `json:"updated_at"`
	Rows      []core.PayrollRow `json:"rows"`
}

func SaveRange(root string, from, to time.Time, rows []core.PayrollRow) error {
	from, to = dateOnly(from), dateOnly(to)
	if from.IsZero() || to.IsZero() || to.Before(from) {
		return fmt.Errorf("invalid cache date range")
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		return err
	}
	byDay := map[string][]core.PayrollRow{}
	for _, r := range rows {
		if r.End.IsZero() {
			continue
		}
		k := dateOnly(r.End).Format("2006-01-02")
		byDay[k] = append(byDay[k], r)
	}
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		s := daySnapshot{Date: key, UpdatedAt: time.Now(), Rows: byDay[key]}
		if s.Rows == nil {
			s.Rows = []core.PayrollRow{}
		}
		if err := writeSnapshotAtomic(filepath.Join(root, key+".json"), s); err != nil {
			return err
		}
	}
	return nil
}

func LoadRange(root string, from, to time.Time) ([]core.PayrollRow, int, error) {
	from, to = dateOnly(from), dateOnly(to)
	if from.IsZero() || to.IsZero() || to.Before(from) {
		return nil, 0, fmt.Errorf("invalid cache date range")
	}
	var rows []core.PayrollRow
	covered := 0
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		s, ok, err := loadDay(root, d)
		if err != nil {
			return nil, covered, err
		}
		if !ok {
			continue
		}
		covered++
		rows = append(rows, s.Rows...)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].End.Equal(rows[j].End) {
			return rows[i].User < rows[j].User
		}
		return rows[i].End.Before(rows[j].End)
	})
	return rows, covered, nil
}

func MissingRanges(root string, from, to, today time.Time) ([]DateRange, error) {
	from, to, today = dateOnly(from), dateOnly(to), dateOnly(today)
	if from.IsZero() || to.IsZero() || to.Before(from) {
		return nil, fmt.Errorf("invalid cache date range")
	}
	var missing []time.Time
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		_, ok, err := loadDay(root, d)
		if err != nil {
			ok = false
		}
		if d.Equal(today) || !ok {
			missing = append(missing, d)
		}
	}
	if len(missing) == 0 {
		return nil, nil
	}
	ranges := []DateRange{{From: missing[0], To: missing[0]}}
	for _, d := range missing[1:] {
		last := &ranges[len(ranges)-1]
		if last.To.AddDate(0, 0, 1).Equal(d) {
			last.To = d
			continue
		}
		ranges = append(ranges, DateRange{From: d, To: d})
	}
	return ranges, nil
}

func RowsForDate(rows []core.PayrollRow, day time.Time) []core.PayrollRow {
	key := dateOnly(day).Format("2006-01-02")
	out := make([]core.PayrollRow, 0)
	for _, r := range rows {
		if !r.End.IsZero() && dateOnly(r.End).Format("2006-01-02") == key {
			out = append(out, r)
		}
	}
	return out
}

func TotalDays(from, to time.Time) int {
	from, to = dateOnly(from), dateOnly(to)
	if from.IsZero() || to.IsZero() || to.Before(from) {
		return 0
	}
	n := 0
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		n++
	}
	return n
}

func writeSnapshotAtomic(path string, s daySnapshot) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0600); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func loadDay(root string, day time.Time) (daySnapshot, bool, error) {
	key := dateOnly(day).Format("2006-01-02")
	path := filepath.Join(root, key+".json")
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return daySnapshot{}, false, nil
	}
	if err != nil {
		return daySnapshot{}, false, err
	}
	var s daySnapshot
	if err := json.Unmarshal(b, &s); err != nil {
		return daySnapshot{}, false, err
	}
	if s.Date != key {
		return daySnapshot{}, false, fmt.Errorf("cache date mismatch")
	}
	if s.Rows == nil {
		s.Rows = []core.PayrollRow{}
	}
	return s, true, nil
}

func dateOnly(t time.Time) time.Time {
	if t.IsZero() {
		return time.Time{}
	}
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
