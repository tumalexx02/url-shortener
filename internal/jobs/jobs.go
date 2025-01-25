package jobs

import "github.com/robfig/cron/v3"

type Job struct {
	Name string
	cron.Job
}
