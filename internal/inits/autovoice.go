package inits

import (
	"github.com/sarulabs/di/v2"
	_ "github.com/sarulabs/di/v2"
	"github.com/zekurio/daemon/internal/services/autovoice"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/util/static"
)

func InitAutovoice(ctn di.Container) *autovoice.AutovoiceHandler {

	db := ctn.Get(static.DiDatabase).(database.Database)

	handler := autovoice.NewAutovoiceHandler(db)

	return handler
}
