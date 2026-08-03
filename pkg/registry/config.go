package registry

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/registry/data"
	"hit.edu/framework/pkg/registry/server"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	DataPath = "./tmp/data"
)

// Config为对外暴露的配置，可以简化
type Config struct {
	ServingInfo *server.ServingInfo

	ShutdownTimeout time.Duration

	MinRequestTimeout time.Duration

	Extra
}

type Extra struct {
	// TODO:
}

func NewConfig() *Config {
	servingOptions := server.NewServingOptions()
	servingInfo, err := NewServingInfo(servingOptions)
	if err != nil {
		logs.Errorf("Failed to create serving info: %v", err)
		return nil
	}

	return &Config{
		ServingInfo:       servingInfo,
		ShutdownTimeout:   60 * time.Second,
		Extra:             Extra{},
		MinRequestTimeout: 1800 * time.Second,
	}
}

func NewServingInfo(s *server.ServingOptions) (*server.ServingInfo, error) {
	if s == nil {
		return nil, fmt.Errorf("serving options is nil")
	}
	if s.BindPort <= 0 && s.Listener == nil {
		return nil, fmt.Errorf("bind port must be greater than 0 when listener is nil")
	}

	if err := os.MkdirAll(filepath.Clean(DataPath), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create data path: %w", err)
	}

	if s.Listener == nil {
		addr := net.JoinHostPort(s.BindAddress.String(), strconv.Itoa(s.BindPort))
		c := net.ListenConfig{}

		listener, bindPort, err := server.CreateListener(s.BindNetwork, addr, c)
		if err != nil {
			return nil, err
		}
		s.Listener = listener
		s.BindPort = bindPort
	}

	dataSpecList := data.NewDataSpecList()
	subscribers := data.NewSubscriptionManager()

	return &server.ServingInfo{
		Listener:     s.Listener,
		DataPath:     DataPath,
		DataSpecList: dataSpecList,
		Subscribers:  subscribers,
		Handlers:     server.NewRegistryHandler(DataPath, dataSpecList, subscribers),
	}, nil
}
