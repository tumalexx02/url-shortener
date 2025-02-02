package database

type Statistic struct {
	TotalURLCount int    `json:"total_url_count" db:"total_url_count"`
	DayPeak       int    `json:"day_peak" db:"day_peak"`
	LeadersJSON   []byte `json:"leaders" db:"leaders"`
}

func NewStatistic(totalUrlCount, dayPeak int, leaders []byte) Statistic {
	return Statistic{
		TotalURLCount: totalUrlCount,
		DayPeak:       dayPeak,
		LeadersJSON:   leaders,
	}
}
