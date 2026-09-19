package core

import (
	"sort"
	"strings"
	"time"
)

type ReportSet struct {
	Recap     Table
	EvenOdd   Table
	NSLDPick  Table
	NSLDPack  Table
}

type reportDay struct {
	Date time.Time

	PickEvenPieces int
	PickOddPieces  int
	PickEvenTime   float64
	PickOddTime    float64
	PickAutoPieces int

	PackEvenPieces int
	PackOddPieces  int
	PackEvenTime   float64
	PackOddTime    float64

	PackNormalPieces int
	PackNormalTime   float64
}

func BuildReports(rows []PayrollRow, from, to time.Time) ReportSet {
	from = dayOnly(from)
	to = dayOnly(to)
	if from.IsZero() || to.IsZero() || to.Before(from) {
		return emptyReports()
	}

	packUserByReference := map[string]string{}
	for _, r := range rows {
		if !isPackJob(r.Job) || !inDateRange(r.End, from, to) {
			continue
		}
		ref := strings.TrimSpace(r.Reference)
		if ref == "" {
			continue
		}
		if _, exists := packUserByReference[ref]; !exists {
			packUserByReference[ref] = strings.TrimSpace(r.User)
		}
	}

	days := map[string]*reportDay{}
	getDay := func(t time.Time) *reportDay {
		d := dayOnly(t)
		key := d.Format("2006-01-02")
		if days[key] == nil {
			days[key] = &reportDay{Date: d}
		}
		return days[key]
	}

	for _, r := range rows {
		if !inDateRange(r.End, from, to) {
			continue
		}
		d := getDay(r.End)
		switch {
		case isPickJob(r.Job):
			if isEven(r.EvenOdd) {
				d.PickEvenPieces += r.Pieces
				d.PickEvenTime += r.Duration
			} else if isOdd(r.EvenOdd) {
				d.PickOddPieces += r.Pieces
				d.PickOddTime += r.Duration
			}
			if ref := strings.TrimSpace(r.Reference); ref != "" && strings.EqualFold(packUserByReference[ref], "robotics") {
				d.PickAutoPieces += r.Pieces
			}
		case isPackJob(r.Job):
			if isEven(r.EvenOdd) {
				d.PackEvenPieces += r.Pieces
				d.PackEvenTime += r.Duration
			} else if isOdd(r.EvenOdd) {
				d.PackOddPieces += r.Pieces
				d.PackOddTime += r.Duration
			}
			if !strings.EqualFold(strings.TrimSpace(r.User), "robotics") {
				d.PackNormalPieces += r.Pieces
				d.PackNormalTime += r.Duration
			}
		}
	}

	keys := make([]string, 0, len(days))
	for k := range days {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := emptyReports()
	for _, k := range keys {
		d := days[k]
		pickTotal := d.PickEvenPieces + d.PickOddPieces
		pickTime := d.PickEvenTime + d.PickOddTime
		totalOutput := d.PackNormalPieces + d.PickAutoPieces
		totalTime := d.PackNormalTime + pickTime

		evenPct, oddPct := 0.0, 0.0
		if pickTotal > 0 {
			evenPct = float64(d.PickEvenPieces) / float64(pickTotal)
			oddPct = float64(d.PickOddPieces) / float64(pickTotal)
		}
		totalNSLD := 0.0
		if totalTime > 0 {
			totalNSLD = float64(totalOutput) / totalTime
		}

		out.Recap.Rows = append(out.Recap.Rows, []any{
			d.Date,
			d.PackNormalTime,
			d.PackNormalPieces,
			d.PickAutoPieces,
			pickTime,
			totalNSLD,
			totalOutput,
			evenPct,
			oddPct,
		})
		out.EvenOdd.Rows = append(out.EvenOdd.Rows, []any{
			d.Date,
			d.PickEvenPieces,
			d.PickOddPieces,
			pickTotal,
			evenPct,
			oddPct,
		})
		out.NSLDPick.Rows = append(out.NSLDPick.Rows, []any{
			d.Date,
			d.PickEvenPieces,
			d.PickOddPieces,
			d.PickEvenTime,
			d.PickOddTime,
			pickTotal,
			pickTime,
			ratioOrZero(float64(d.PickEvenPieces), d.PickEvenTime),
			ratioOrZero(float64(d.PickOddPieces), d.PickOddTime),
		})
		packTotal := d.PackEvenPieces + d.PackOddPieces
		packTime := d.PackEvenTime + d.PackOddTime
		out.NSLDPack.Rows = append(out.NSLDPack.Rows, []any{
			d.Date,
			d.PackEvenPieces,
			d.PackEvenTime,
			d.PackOddPieces,
			d.PackOddTime,
			packTotal,
			packTime,
			ratioOrZero(float64(d.PackEvenPieces), d.PackEvenTime),
			ratioOrZero(float64(d.PackOddPieces), d.PackOddTime),
		})
	}
	return out
}

func emptyReports() ReportSet {
	return ReportSet{
		Recap: Table{Headers: []string{
			"Ngày", "Time Pack hàng thường", "SL Pack hàng thường", "SL Pick Auto PP",
			"Time Pick", "NSLD tổng", "Tổng SL Hàng thường + Auto PP", "% Chẵn Pick", "% Lẻ Pick",
		}},
		EvenOdd: Table{Headers: []string{
			"Ngày", "SL Chẵn", "SL Lẻ", "Tổng SL", "% Chẵn", "% Lẻ",
		}},
		NSLDPick: Table{Headers: []string{
			"Ngày", "SL Chẵn", "SL Lẻ", "Time Chẵn", "Time Lẻ", "Tổng SL", "Tổng Time", "NSLD Pick Chẵn", "NSLD Pick Lẻ",
		}},
		NSLDPack: Table{Headers: []string{
			"Ngày", "SL Chẵn", "Time Chẵn", "SL Lẻ", "Time Lẻ", "Tổng SL", "Tổng Time", "NSLD Pack Chẵn", "NSLD Pack Lẻ",
		}},
	}
}

func ratioOrZero(n, d float64) float64 {
	if d <= 0 {
		return 0
	}
	return n / d
}

func dayOnly(t time.Time) time.Time {
	if t.IsZero() {
		return time.Time{}
	}
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func inDateRange(t, from, to time.Time) bool {
	if t.IsZero() {
		return false
	}
	d := dayOnly(t)
	return !d.Before(from) && !d.After(to)
}
