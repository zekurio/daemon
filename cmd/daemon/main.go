package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/zekrotja/ken"
	"github.com/zekurio/daemon/internal/inits"
	"github.com/zekurio/daemon/internal/models"

	"github.com/charmbracelet/log"
	"github.com/sarulabs/di/v2"
	"github.com/zekurio/daemon/internal/services/config"
	"github.com/zekurio/daemon/internal/services/database"
	"github.com/zekurio/daemon/internal/services/permissions"
	"github.com/zekurio/daemon/internal/util/embedded"
	"github.com/zekurio/daemon/internal/util/static"
)

var (
	flagConfigPath = flag.String("c", "config.toml", "Path to config file")
)

func main() {

	flag.Parse()

	if embedded.Release == "true" {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.DebugLevel)
	}

	diBuilder, err := di.NewBuilder()
	if err != nil {
		log.With(err).Fatal("Failed to create DI builder")
	}

	// Config
	err = diBuilder.Add(di.Def{
		Name: static.DiConfig,
		Build: func(ctn di.Container) (interface{}, error) {
			return config.Parse(*flagConfigPath, "DAEMON_", models.DefaultConfig)
		},
	})
	if err != nil {
		log.With(err).Fatal("Config parsing failed")
	}

	// Database
	err = diBuilder.Add(di.Def{
		Name: static.DiDatabase,
		Build: func(ctn di.Container) (interface{}, error) {
			return inits.InitDatabase(ctn)
		},
		Close: func(obj interface{}) error {
			d := obj.(database.Database)
			log.Info("Shutting down database connection...")
			err := d.Close()
			if err != nil {
				return err
			}
			return nil
		},
	})
	if err != nil && err.Error() == "unknown database driver" {
		log.With(err).Fatal("Database creation failed, unknown driver")
	} else if err != nil {
		log.With(err).Fatal("Database creation failed")
	}

	// Permissions
	err = diBuilder.Add(di.Def{
		Name: static.DiPermissions,
		Build: func(ctn di.Container) (interface{}, error) {
			return permissions.InitPermissions(ctn), nil
		},
	})
	if err != nil {
		log.With(err).Fatal("Permissions creation failed")
	}

	// Discord Session
	err = diBuilder.Add(di.Def{
		Name: static.DiDiscord,
		Build: func(ctn di.Container) (interface{}, error) {
			return inits.InitDiscord(ctn)
		},
		Close: func(obj interface{}) error {
			return obj.(*discordgo.Session).Close()
		},
	})
	if err != nil {
		log.With(err).Fatal("Discord creation failed")
	}

	// Ken
	err = diBuilder.Add(di.Def{
		Name: static.DiCommandHandler,
		Build: func(ctn di.Container) (interface{}, error) {
			return inits.InitKen(ctn)
		},
		Close: func(obj interface{}) error {
			return obj.(*ken.Ken).Unregister()
		},
	})
	if err != nil {
		log.With(err).Fatal("Command handler creation failed")
	}

	// Scheduler
	err = diBuilder.Add(di.Def{
		Name: static.DiScheduler,
		Build: func(ctn di.Container) (interface{}, error) {
			return inits.InitScheduler(ctn), nil
		},
	})
	if err != nil {
		log.With(err).Fatal("Scheduler creation failed")
	}

	// Build dependency injection container
	ctn := diBuilder.Build()
	// Tear down dependency instances
	defer func(ctn di.Container) {
		err := ctn.DeleteWithSubContainers()
		if err != nil {
			log.With(err).Fatal("Failed to tear down dependency instances")
		}
	}(ctn)

	ctn.Get(static.DiCommandHandler)

	s := ctn.Get(static.DiDiscord).(*discordgo.Session)
	err = s.Open()
	if err != nil {
		log.With(err).Fatal("Failed to open Discord connection")
	}

	ctn.Get(static.DiDatabase)

	// Block main go routine until one of the following
	// specified exit sys calls occure.
	log.Info("Started event loop. Stop with CTRL-C...")

	log.Info("Initialization finished")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

}
