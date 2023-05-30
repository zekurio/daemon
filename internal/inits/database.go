package inits

import (
	"github.com/charmbracelet/log"
	"github.com/sarulabs/di/v2"
	"github.com/zekurio/daemon/internal/models"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/services/database/postgres"
	"github.com/zekurio/daemon/internal/util/static"
)

func InitDatabase(ctn di.Container) (database.Database, error) {
	var db database.Database
	var err error

	cfg := ctn.Get(static.DiConfig).(models.Config)
	db, err = postgres.InitPostgres(cfg.Postgres)

	if err != nil {
		return nil, err
	}

	log.Info("Connected to database")

	return db, nil
}
