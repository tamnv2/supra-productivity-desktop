package core

import (
	"math"
	"testing"
	"time"
)

func TestBuildReportsMatchesRecapWorkbookLogic(t *testing.T) {
	day := time.Date(2026, 9, 11, 0, 0, 0, 0, time.Local)
	at := func(h int) time.Time { return time.Date(2026, 9, 11, h, 0, 0, 0, time.Local) }
	rows := []PayrollRow{
		{Job: "Pick", EvenOdd: "Chẵn", Reference: "REF-A", User: "picker1", Start: at(8), End: at(9), Duration: 7320, Pieces: 138929},
		{Job: "Pick", EvenOdd: "Lẻ", Reference: "REF-B", User: "picker2", Start: at(9), End: at(10), Duration: 21179, Pieces: 120137},
		{Job: "Pack", EvenOdd: "Chẵn", Reference: "REF-A", User: "robotics", Start: at(8), End: at(9), Duration: 274, Pieces: 138929},
		{Job: "Pack", EvenOdd: "Lẻ", Reference: "REF-B", User: "pack1", Start: at(9), End: at(10), Duration: 7339, Pieces: 127406},
	}
	r := BuildReports(rows, day, day)
	if len(r.Recap.Rows) != 1 {
		t.Fatalf("recap rows=%d", len(r.Recap.Rows))
	}
	got := r.Recap.Rows[0]
	if got[1] != float64(7339) || got[2] != 127406 || got[3] != 138929 || got[4] != float64(28499) || got[6] != 266335 {
		t.Fatalf("unexpected recap row: %#v", got)
	}
	wantNSLD := float64(266335) / float64(7339+28499)
	if math.Abs(got[5].(float64)-wantNSLD) > 1e-12 {
		t.Fatalf("NSLD=%v want %v", got[5], wantNSLD)
	}
	wantEvenPct := float64(138929) / float64(138929+120137)
	if math.Abs(got[7].(float64)-wantEvenPct) > 1e-12 {
		t.Fatalf("even%%=%v want %v", got[7], wantEvenPct)
	}
	if len(r.EvenOdd.Rows) != 1 || r.EvenOdd.Rows[0][3] != 259066 {
		t.Fatalf("even/odd=%#v", r.EvenOdd.Rows)
	}
	if len(r.NSLDPick.Rows) != 1 || r.NSLDPick.Rows[0][6] != float64(28499) {
		t.Fatalf("pick NSLD=%#v", r.NSLDPick.Rows)
	}
	if len(r.NSLDPack.Rows) != 1 || r.NSLDPack.Rows[0][5] != 266335 || r.NSLDPack.Rows[0][6] != float64(7613) {
		t.Fatalf("pack NSLD=%#v", r.NSLDPack.Rows)
	}
}

func TestBuildReportsDoesNotTreatBlankReferenceAsAutoPP(t *testing.T) {
	day := time.Date(2026, 9, 11, 0, 0, 0, 0, time.Local)
	rows := []PayrollRow{
		{Job: "Pick", EvenOdd: "Chẵn", Reference: "", User: "picker", End: day.Add(9 * time.Hour), Duration: 10, Pieces: 100},
		{Job: "Pack", EvenOdd: "Chẵn", Reference: "", User: "robotics", End: day.Add(9 * time.Hour), Duration: 10, Pieces: 100},
	}
	r := BuildReports(rows, day, day)
	if got := r.Recap.Rows[0][3]; got != 0 {
		t.Fatalf("blank reference must not map Auto PP, got=%v", got)
	}
}


func TestBuildReportsUsesFirstPackMatchLikeVLookup(t *testing.T) {
	day := time.Date(2026, 9, 11, 0, 0, 0, 0, time.Local)
	rows := []PayrollRow{
		{Job: "Pick", EvenOdd: "Chẵn", Reference: "REF-X", User: "picker", End: day.Add(9 * time.Hour), Duration: 10, Pieces: 100},
		{Job: "Pack", EvenOdd: "Chẵn", Reference: "REF-X", User: "pack-normal", End: day.Add(9 * time.Hour), Duration: 10, Pieces: 100},
		{Job: "Pack", EvenOdd: "Chẵn", Reference: "REF-X", User: "robotics", End: day.Add(10 * time.Hour), Duration: 10, Pieces: 100},
	}
	r := BuildReports(rows, day, day)
	if got := r.Recap.Rows[0][3]; got != 0 {
		t.Fatalf("first Pack match is not robotics, Auto PP must be 0; got=%v", got)
	}
}
