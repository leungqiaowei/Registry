package server

import (
	"context"
	"fmt"
	"hit.edu/framework/pkg/registry/data"
	"net"
)

// 配置运行时的Serving信息
// Serving信息应该从Option中获取
type ServingInfo struct {
	//
	Listener net.Listener

	// 数据存储位置
	DataPath string

	// file mapping
	//FileMapping *data.FileMapping
	DataSpecList *data.DataSpecList

	// 订阅者列表
	Subscribers *data.SubscriptionManager

	// 各类Handler
	Handlers *RegistryHandler

	// TODO: 配置HTTP相关参数
}

func CreateListener(network, addr string, config net.ListenConfig) (net.Listener, int, error) {
	if len(network) == 0 {
		network = "tcp"
	}

	ln, err := config.Listen(context.TODO(), network, addr)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to listen on %v: %v", addr, err)
	}

	// get port
	tcpAddr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		ln.Close()
		return nil, 0, fmt.Errorf("invalid listen address: %q", ln.Addr().String())
	}

	return ln, tcpAddr.Port, nil
}

type ServingOptions struct {
	// 连接地址
	BindAddress net.IP
	// HTTP连接端口
	BindPort int
	// HTTP连接网络,tcp4 or tcp6
	BindNetwork string
	// TODO: 本地Hostname
	BindHost string
	// 对外广播的地址，用于服务发现
	ExternalAddress net.IP
	// HTTP Lisntener
	Listener net.Listener
}

// 默认配置
func NewServingOptions() *ServingOptions {
	return &ServingOptions{
		BindAddress: net.ParseIP("0.0.0.0"),
		BindPort:    8119,
	}
}
