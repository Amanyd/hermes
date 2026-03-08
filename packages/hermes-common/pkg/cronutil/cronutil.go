package cronutil

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// Returns the next scheduled time after `from` for the given
// standard 5-field cron expression (minute hour dom month dow).
func ComputeNextRun(schedule string, from time.Time) (time.Time, error) {
	if schedule == "" {
		return time.Time{}, fmt.Errorf("cron schedule is empty")
	}
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	sched, err := parser.Parse(schedule)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse cron %q: %w", schedule, err)
	}
	return sched.Next(from), nil
}
