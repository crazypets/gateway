package app

import (
	"context"
	"fmt"
	"gateway/internal/config"
	"gateway/internal/infrastructure/consul"
	httptransport "gateway/internal/transport/http"
	consulapi "github.com/hashicorp/consul/api"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/sync/errgroup"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const httpTimeoutClose = 5 * time.Second

type Option func(app *App)

func WithLogger(logger *zap.Logger) Option {
	return func(app *App) {
		app.logger = logger
	}
}

// WithServiceID adding service id option for consul register.
func WithServiceID(id string) Option {
	return func(app *App) {
		app.svcID = id
	}
}

type App struct {
	cfg   *config.Config
	svcID string

	consulSvc *consul.Service
	httpSrv   *http.Server

	logger *zap.Logger
}

func New(cfg *config.Config, opts ...Option) *App {
	app := &App{
		cfg:    cfg,
		svcID:  uuid.NewV4().String(),
		logger: zap.NewNop(),
	}

	for _, opt := range opts {
		opt(app)
	}

	return app
}

func (a *App) Run() error {
	eg := errgroup.Group{}

	a.httpSrv = httptransport.NewHTTPServer(a.cfg)

	consulClient, err := a.initConsul(a.cfg.Consul.Addr)
	if err != nil {
		return err
	}

	a.consulSvc = consul.NewService(consulClient, a.cfg.Consul, a.svcID)

	if err = a.consulSvc.Register(); err != nil {
		return err
	}

	eg.Go(func() error {
		if err = a.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("http serve error: %w", err)
		}

		return nil
	})

	return eg.Wait()
}

func (a *App) initConsul(addr string) (*consulapi.Client, error) {
	consulCfg := consulapi.DefaultConfig()
	consulCfg.Address = addr

	client, err := consulapi.NewClient(consulCfg)
	if err != nil {
		return nil, fmt.Errorf("create consul client: %w", err)
	}

	return client, nil
}

func (a *App) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), httpTimeoutClose)
	defer cancel()

	if err := a.consulSvc.Deregister(); err != nil {
		return err
	}

	if err := a.httpSrv.Shutdown(ctx); err != nil {
		return fmt.Errorf("http closing error: %w", err)
	}

	return nil
}
