package scheduler

import (
	"errors"

	"github.com/robfig/cron/v3"
)

type CronScheduler struct {
	C *cron.Cron
}

var _ Provider = (*CronScheduler)(nil)

func (c *CronScheduler) Schedule(spec any, job func()) (id any, err error) {
	s, ok := spec.(string)
	if !ok {
		return nil, errors.New("spec must be a string")
	}

	return c.C.AddFunc(s, job)
}

func (c *CronScheduler) Unschedule(id any) error {
	i, ok := id.(cron.EntryID)
	if !ok {
		return errors.New("id must be a cron.EntryID")
	}

	c.C.Remove(i)
	return nil
}

func (c *CronScheduler) Start() {
	c.C.Start()
}

func (c *CronScheduler) Stop() {
	c.C.Stop()
}
