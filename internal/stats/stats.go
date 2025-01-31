package stats

type Statistic struct {
	TotalURLCount int `json:"total_url_count"`
	UrlPerMinute  int `json:"url_per_minute"`
	DayPeak       int `json:"day_peak"`
}

func NewStatistic(totalUrlCount, urlPerMinute, dayPeak int) Statistic {
	return Statistic{
		TotalURLCount: totalUrlCount,
		UrlPerMinute:  urlPerMinute,
		DayPeak:       dayPeak,
	}
}
