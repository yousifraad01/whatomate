package handlers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFillDailySeries_FillsGapsWithZero(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC)

	got := fillDailySeries(map[string]float64{"2026-09-01": 3, "2026-09-03": 1}, start, end)

	want := []ChartPoint{
		{Label: "Sep 01", Value: 3},
		{Label: "Sep 02", Value: 0},
		{Label: "Sep 03", Value: 1},
		{Label: "Sep 04", Value: 0},
	}
	assert.Equal(t, want, got)
}

func TestFillDailySeries_LongRangeKeepsOnlyDaysWithData(t *testing.T) {
	t.Parallel()

	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	got := fillDailySeries(map[string]float64{"2025-12-31": 2, "2021-06-15": 5}, start, end)

	want := []ChartPoint{
		{Label: "Jun 15", Value: 5},
		{Label: "Dec 31", Value: 2},
	}
	assert.Equal(t, want, got, "sparse days stay in date order without zero-fill")
}

func TestFillDailySeries_EmptyRange(t *testing.T) {
	t.Parallel()

	day := time.Date(2026, 9, 1, 8, 0, 0, 0, time.UTC)
	got := fillDailySeries(map[string]float64{}, day, day)
	assert.Equal(t, []ChartPoint{{Label: "Sep 01", Value: 0}}, got)
}
