package stats

import "time"

type Statistic struct {
	TotalURLCount int            `json:"total_url_count" db:"total_url_count"`
	UrlPerMinute  int            `json:"url_per_min" db:"url_per_min"`
	DayPeak       int            `json:"day_peak" db:"day_peak"`
	Leaders       []ResourceInfo `json:"leaders" db:"leaders"`
}

type DayPeakStatistic struct {
	DayPeak    int       `db:"day_peak"`
	LastUpdate time.Time `db:"updated_at"`
}

type ResourceInfo struct {
	Resource string `json:"resource" db:"resource"`
	URLCount int    `json:"url_count" db:"url_count"`
}

func NewStatistic(totalUrlCount, urlPerMinute, dayPeak int, leaders []ResourceInfo) Statistic {
	return Statistic{
		TotalURLCount: totalUrlCount,
		UrlPerMinute:  urlPerMinute,
		DayPeak:       dayPeak,
		Leaders:       leaders,
	}
}
