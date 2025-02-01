package stats

import "time"

type Statistic struct {
	TotalURLCount int `json:"total_url_count" db:"total_url_count"`
	UrlPerMinute  int `json:"url_per_min" db:"url_per_min"`
	DayPeak       int `json:"day_peak" db:"day_peak"`
}

type DayPeakStatistic struct {
	DayPeak    int       `json:"day_peak" db:"day_peak"`
	LastUpdate time.Time `json:"last_update" db:"timestamp"`
}

func NewStatistic(totalUrlCount, urlPerMinute, dayPeak int) Statistic {
	return Statistic{
		TotalURLCount: totalUrlCount,
		UrlPerMinute:  urlPerMinute,
		DayPeak:       dayPeak,
	}
}
