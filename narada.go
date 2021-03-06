package narada

import (
	"context"

	"github.com/m1ome/narada/clients"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type (
	Narada struct {
		providers []interface{}
		logger    *logrus.Logger
		config    *viper.Viper
		app       *fx.App
	}
)

func (t Narada) HandleError(err error) {
	t.logger.Fatal(err)
}

func New(name string, version string, providers ...interface{}) *Narada {
	config, err := NewConfig()
	if err != nil {
		logger, _ := NewLogger(viper.New())
		logger.WithField("error", err).Fatal("error reading configuration")
	}

	config.SetDefault("app.name", name)
	config.SetDefault("app.version", version)

	logger, err := NewLogger(config)
	if err != nil {
		logger, _ := NewLogger(viper.New())
		logger.WithField("error", err).Fatal("error creating logger from configuration")
	}

	return &Narada{
		providers: providers,
		logger:    logger,
		config:    config,
	}
}

func (t *Narada) Start(fn interface{}) {
	// Creating application
	t.app = fx.New(
		// Setting default logger to discard
		fx.Logger(NewNopLogger()),

		fx.ErrorHook(t),

		fx.Provide(
			// Fundamentals
			NewConfig,
			NewSentry,
			NewLogger,

			// Servers handling
			NewMultiServers,

			// Workers handling
			NewWorkers,

			// Clients
			clients.NewPostgreSQL,
			clients.NewRedis,
		),

		fx.Provide(t.providers...),

		fx.Invoke(
			// Adding servers by default
			NewMetricsInvoke,
			NewProfilerInvoke,

			// Invoke user-defined function
			fn,
		),
	)

	t.app.Run()
}

func (t *Narada) Stop() {
	t.app.Stop(context.Background())
}
