package scheduler

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

// Job is one scheduled task.
type Job struct {
	Label string
	Spec  string
	Run   func() error
}

// Scheduler wraps a cron instance and exposes a small control surface for the
// TUI and the headless daemon.
type Scheduler struct {
	cron    *cron.Cron
	logger  *slog.Logger
	entryID map[string]cron.EntryID
	running bool
}

// New creates a scheduler from the given jobs. Each job's spec is validated
// before the scheduler is returned.
func New(jobs []Job, logger *slog.Logger) (*Scheduler, error) {
	s := &Scheduler{
		logger:  logger,
		entryID: map[string]cron.EntryID{},
	}
	c := cron.New()

	for _, job := range jobs {
		if job.Label == "" {
			return nil, fmt.Errorf("job label is required")
		}
		if _, err := cron.ParseStandard(job.Spec); err != nil {
			return nil, fmt.Errorf("parsing spec for %s: %w", job.Label, err)
		}

		spec := job.Spec
		run := job.Run
		entryID, err := c.AddFunc(spec, func() {
			if err := run(); err != nil {
				s.logger.Error("scheduled job failed", "job", job.Label, "error", err)
			}
		})
		if err != nil {
			return nil, fmt.Errorf("scheduling %s: %w", job.Label, err)
		}
		s.entryID[job.Label] = entryID
	}

	s.cron = c
	return s, nil
}

func (s *Scheduler) Start() error {
	if s.running {
		return nil
	}
	s.cron.Start()
	s.running = true
	return nil
}

func (s *Scheduler) Stop() {
	if !s.running {
		return
	}
	s.cron.Stop()
	s.running = false
}

func (s *Scheduler) Running() bool {
	return s.running
}

// NextRuns returns the next fire time for each job label. Times are zero when
// the scheduler is not running.
func (s *Scheduler) NextRuns() map[string]time.Time {
	out := map[string]time.Time{}
	for label, id := range s.entryID {
		if entry := s.cron.Entry(id); entry.Valid() {
			out[label] = entry.Next
		}
	}
	return out
}

// DailySpec builds a cron expression that fires every day at HH:MM.
func DailySpec(hhmm string) (string, error) {
	minute, hour, err := parseTime(hhmm)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d %d * * *", minute, hour), nil
}

// WeeklySpec builds a cron expression that fires on a named weekday at HH:MM.
func WeeklySpec(day, hhmm string) (string, error) {
	minute, hour, err := parseTime(hhmm)
	if err != nil {
		return "", err
	}
	dow, ok := weekdays[strings.ToLower(strings.TrimSpace(day))]
	if !ok {
		return "", fmt.Errorf("unknown weekday %q", day)
	}
	return fmt.Sprintf("%d %d * * %d", minute, hour, dow), nil
}

var weekdays = map[string]int{
	"sunday": 0, "monday": 1, "tuesday": 2, "wednesday": 3,
	"thursday": 4, "friday": 5, "saturday": 6,
}

func parseTime(hhmm string) (minute, hour int, err error) {
	parts := strings.SplitN(strings.TrimSpace(hhmm), ":", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid time %q, want HH:MM", hhmm)
	}

	hour, err = strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || hour < 0 || hour > 23 {
		return 0, 0, fmt.Errorf("invalid hour in %q", hhmm)
	}
	minute, err = strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || minute < 0 || minute > 59 {
		return 0, 0, fmt.Errorf("invalid minute in %q", hhmm)
	}
	return minute, hour, nil
}
