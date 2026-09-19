package core

import (
	"testing"
	"time"
)

func TestRecoveredV13Defaults(t *testing.T) {
	b := DefaultBusinessSettings()
	if b.ActiveStatus != "Tất cả" || b.PickShift != "Ca 2" || b.PackShift != "Tất cả" {
		t.Fatalf("unexpected filters: %#v", b)
	}
	if b.PickEvenQuota != 8 || b.PackEvenQuota != 5 {
		t.Fatalf("unexpected quotas: pick=%d pack=%d", b.PickEvenQuota, b.PackEvenQuota)
	}
	if b.PickTargetInhouseEven != 8550 || b.PickTargetInhouseOdd != 2250 ||
		b.PickTargetOtherEven != 10000 || b.PickTargetOtherOdd != 2100 {
		t.Fatalf("unexpected recovered targets: %#v", b)
	}
	if b.PickSpeedInhouseEven != 17 || b.PickSpeedInhouseOdd != 7 ||
		b.PickSpeedOtherEven != 17 || b.PickSpeedOtherOdd != 7 {
		t.Fatalf("unexpected recovered speed settings: %#v", b)
	}
	if b.CheckStart != "07:00" || b.CheckEnd != "12:48" || b.CheckLimit != 150 {
		t.Fatalf("unexpected 1C1L window: %#v", b)
	}
}

func TestManualShiftWins(t *testing.T) {
	b := DefaultBusinessSettings()
	b.PickShift = "Tất cả"
	b.ManualShifts[ManualShiftKey("u1", "Pick")] = "Ca 2"
	rows := []PayrollRow{{
		Job: "Pick", EvenOdd: "Chẵn", DO: "1", User: "u1", Name: "A",
		Provider: "Inhouse", Site: "1921",
		Start: time.Date(2026, 9, 1, 8, 0, 0, 0, time.Local),
		End: time.Date(2026, 9, 1, 9, 0, 0, 0, time.Local),
		Pieces: 100, Duration: 10,
	}}
	_, _, s := BuildTables(rows, b, time.Date(2026, 9, 1, 10, 0, 0, 0, time.Local))
	if len(s.Rows) != 1 || s.Rows[0][6] != "Ca 2" {
		t.Fatalf("manual shift not applied: %#v", s.Rows)
	}
}

func TestRecoveredV13PerDODeduction(t *testing.T) {
	b := DefaultBusinessSettings()
	b.PickShift = "Tất cả"
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.Local)
	rows := []PayrollRow{
		{
			Job: "Pick", EvenOdd: "Chẵn", DO: "DO-1", User: "u1", Name: "A",
			Provider: "Inhouse", Site: "1921",
			Start: now.Add(-2 * time.Hour), End: now.Add(-90 * time.Minute),
			Pieces: 100, SKU: 11, Duration: 6,
		},
		{
			Job: "Pick", EvenOdd: "Chẵn", DO: "DO-2", User: "u1", Name: "A",
			Provider: "Inhouse", Site: "1921",
			Start: now.Add(-80 * time.Minute), End: now.Add(-60 * time.Minute),
			Pieces: 100, SKU: 21, Duration: 6,
		},
	}
	pick, _, _ := BuildTables(rows, b, now)
	if len(pick.Rows) != 1 {
		t.Fatalf("pick rows=%d", len(pick.Rows))
	}
	// V1.3: 200 pieces - ceil(11/10) - ceil(21/10) = 195.
	if got := pick.Rows[0][8]; got != 195 {
		t.Fatalf("V1.3 per-DO SKU deduction mismatch: got=%v", got)
	}
}

func TestRecoveredV13SpeedUsesMinutesAndFloors(t *testing.T) {
	b := DefaultBusinessSettings()
	b.PickShift = "Tất cả"
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.Local)
	rows := []PayrollRow{{
		Job: "Pick", EvenOdd: "Chẵn", DO: "DO-1", User: "u1", Name: "A",
		Provider: "Inhouse", Site: "1921",
		Start: now.Add(-2 * time.Hour), End: now.Add(-time.Hour),
		Pieces: 100, SKU: 0, Duration: 6,
	}}
	pick, _, _ := BuildTables(rows, b, now)
	if len(pick.Rows) != 1 {
		t.Fatalf("pick rows=%d", len(pick.Rows))
	}
	if got := pick.Rows[0][18]; got != 16 {
		t.Fatalf("speed must be int(100/6)=16, got=%v", got)
	}
}

func TestRecoveredV13RoboticsAlwaysCa1(t *testing.T) {
	b := DefaultBusinessSettings()
	b.PickShift = "Tất cả"
	now := time.Date(2026, 9, 1, 20, 0, 0, 0, time.Local)
	rows := []PayrollRow{{
		Job: "Pick", EvenOdd: "Lẻ", DO: "DO-1", User: "robot1", Name: "Robot",
		Provider: "Robotics", Site: "1921",
		Start: now.Add(-2 * time.Hour), End: now.Add(-time.Hour),
		Pieces: 10, Duration: 10,
	}}
	_, _, shift := BuildTables(rows, b, now)
	if len(shift.Rows) != 1 || shift.Rows[0][5] != "Ca 1" {
		t.Fatalf("robotics auto shift mismatch: %#v", shift.Rows)
	}
}

func TestParseTenureDays(t *testing.T) {
	if ParseTenureDays("1 năm 2 tháng 3 ngày") <= ParseTenureDays("11 tháng 29 ngày") {
		t.Fatal("tenure must sort by duration")
	}
}

func TestPercentType(t *testing.T) {
	if FormatPercent(0.42) != "42.0%" {
		t.Fatal(FormatPercent(0.42))
	}
}
