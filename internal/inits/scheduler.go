package inits

import (
	"github.com/robfig/cron/v3"
	"github.com/sarulabs/di/v2"
	"github.com/zekurio/daemon/internal/services/scheduler"
)

func InitScheduler(ctn di.Container) scheduler.Provider {

	sched := &scheduler.CronScheduler{C: cron.New(cron.WithSeconds())}

	return sched

}
