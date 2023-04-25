package inits

import (
	"github.com/robfig/cron/v3"
	"github.com/sarulabs/di/v2"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/services/scheduler"
	"github.com/zekurio/daemon/internal/util/static"
)

func InitScheduler(ctn di.Container) scheduler.Provider {

	_ = ctn.Get(static.DiDatabase).(database.Database)

	sched := &scheduler.CronScheduler{C: cron.New(cron.WithSeconds())}

	return sched

}
