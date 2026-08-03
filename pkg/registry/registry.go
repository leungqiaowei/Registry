package registry

import (
	"context"
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/registry/server"
	"net/http"
	"time"
)

const (
	RegistryName = "registry"
)

type Registry struct {
	RegistryInterface

	ServingInfo *server.ServingInfo

	ShutdownTimeout time.Duration

	minRequestTimeout time.Duration
}

type RegistryInterface interface {
	Run(ctx context.Context) error
	Destroy()
}

func NewRegistry(cfg *Config) (*Registry, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if cfg.ServingInfo == nil || cfg.ServingInfo.Listener == nil {
		return nil, fmt.Errorf("serving info is incomplete")
	}

	return &Registry{
		ServingInfo:       cfg.ServingInfo,
		ShutdownTimeout:   cfg.ShutdownTimeout,
		minRequestTimeout: cfg.MinRequestTimeout,
	}, nil
}

func (r *Registry) Destroy() {
	if r == nil || r.ServingInfo == nil || r.ServingInfo.Listener == nil {
		return
	}
	if err := r.ServingInfo.Listener.Close(); err != nil {
		logs.Errorf("failed to close listener: %v", err)
	}
}

func (r *Registry) Run(ctx context.Context) error {
	logs.Info("Running Registry")

	serverInstance := &http.Server{
		Handler:           nil,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := serverInstance.Serve(r.ServingInfo.Listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), r.ShutdownTimeout)
		defer cancel()
		if err := serverInstance.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("failed to shutdown registry: %w", err)
		}
		logs.Info("Stopping Registry")
		return nil
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("failed to start server: %w", err)
		}
		logs.Info("Stopping Registry")
		return nil
	}
}
