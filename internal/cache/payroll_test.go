package cache

import (
	"testing"
	"time"

	"github.com/tamnv2/supra-productivity-desktop/internal/core"
)

func TestMissingRangesFetchesOnlyGapsAndAlwaysRefreshesToday(t *testing.T) {
	root := t.TempDir()
	loc := time.Local
	from := time.Date(2026, 9, 1, 0, 0, 0, 0, loc)
	today := time.Date(2026, 9, 19, 0, 0, 0, 0, loc)
	if err := SaveRange(root, from, time.Date(2026, 9, 15, 0, 0, 0, 0, loc), nil); err != nil {
		t.Fatal(err)
	}
	ranges, err := MissingRanges(root, from, today, today)
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 1 || !ranges[0].From.Equal(time.Date(2026, 9, 16, 0, 0, 0, 0, loc)) || !ranges[0].To.Equal(today) {
		t.Fatalf("unexpected ranges: %#v", ranges)
	}
	if err := SaveRange(root, time.Date(2026, 9, 16, 0, 0, 0, 0, loc), today, nil); err != nil {
		t.Fatal(err)
	}
	ranges, err = MissingRanges(root, from, today, today)
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 1 || !ranges[0].From.Equal(today) || !ranges[0].To.Equal(today) {
		t.Fatalf("today must refresh: %#v", ranges)
	}
}

func TestEmptyHistoricalDayCountsAsCached(t *testing.T) {
	root := t.TempDir()
	d := time.Date(2026, 9, 10, 0, 0, 0, 0, time.Local)
	if err := SaveRange(root, d, d, nil); err != nil {
		t.Fatal(err)
	}
	ranges, err := MissingRanges(root, d, d, d.AddDate(0, 0, 1))
	if err != nil {
		t.Fatal(err)
	}
	if len(ranges) != 0 {
		t.Fatalf("empty successful day must be cached: %#v", ranges)
	}
	rows, covered, err := LoadRange(root, d, d)
	if err != nil || covered != 1 || len(rows) != 0 {
		t.Fatalf("rows=%d covered=%d err=%v", len(rows), covered, err)
	}
}

func TestRowsForDate(t *testing.T) {
	d1 := time.Date(2026, 9, 10, 0, 0, 0, 0, time.Local)
	d2 := d1.AddDate(0, 0, 1)
	rows := []core.PayrollRow{
		{User: "a", End: d1.Add(8 * time.Hour)},
		{User: "b", End: d2.Add(8 * time.Hour)},
	}
	got := RowsForDate(rows, d2)
	if len(got) != 1 || got[0].User != "b" {
		t.Fatalf("got=%#v", got)
	}
}
