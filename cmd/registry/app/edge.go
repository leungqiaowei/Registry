package app

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/component-base/version"
	"hit.edu/framework/pkg/registry/edge"
	"hit.edu/framework/pkg/server"
)

var (
	edgeID            string
	cloudURL          string
	localBaseURL      string
	edgeToken         string
	reconnectInterval time.Duration
	heartbeatInterval time.Duration
	requestTimeout    time.Duration
)

func NewEdgeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edge",
		Short: "启动边侧 Agent，主动连接云端建立云边反向隧道",
		Long: `边侧 Agent 主动连接云端公网 IP 的 /edge/ws 接口，
建立 WebSocket 长连接后，云端可通过 /edges/{edge_id}/{target_path} 反向访问本边侧内网服务。

示例:
  registry edge --cloud-url=ws://120.220.95.189:8119/edge/ws --edge-id=edge-a
  registry edge --cloud-url=ws://cloud.example.com:48119/edge/ws --edge-id=factory-001 --local-base-url=http://127.0.0.1:8080`,
		RunE: runEdge,
	}

	cmd.Flags().StringVar(&edgeID, "edge-id", "", "边侧节点唯一标识（不传则自动使用主机名）")
	cmd.Flags().StringVar(&cloudURL, "cloud-url", "", "云端公网 WebSocket 地址，如 ws://120.220.95.189:8119/edge/ws（必填）")
	cmd.Flags().StringVar(&localBaseURL, "local-base-url", "http://127.0.0.1:8119", "边侧本地要代理的服务基地址")
	cmd.Flags().StringVar(&edgeToken, "token", "", "云端鉴权 Token（可选）")
	cmd.Flags().DurationVar(&reconnectInterval, "reconnect-interval", 5*time.Second, "断线重连间隔")
	cmd.Flags().DurationVar(&heartbeatInterval, "heartbeat-interval", 30*time.Second, "心跳间隔")
	cmd.Flags().DurationVar(&requestTimeout, "request-timeout", 25*time.Second, "边侧请求本地服务的超时时间")

	return cmd
}

func runEdge(cmd *cobra.Command, args []string) error {
	if cloudURL == "" {
		return fmt.Errorf("--cloud-url is required, e.g. ws://120.220.95.189:8119/edge/ws")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		stopCh := server.SetupSignalHandler()
		<-stopCh
		cancel()
	}()

	logs.Init("registry-edge")
	logs.Infof("Starting Edge Agent, version %s", version.Get())
	logs.Infof("Edge ID: %s", edgeID)
	logs.Infof("Cloud URL: %s", cloudURL)
	logs.Infof("Local Base URL: %s", localBaseURL)

	agent := edge.NewAgent(edge.AgentConfig{
		EdgeID:            edgeID,
		CloudURL:          cloudURL,
		LocalBaseURL:      localBaseURL,
		Token:             edgeToken,
		ReconnectInterval: reconnectInterval,
		HeartbeatInterval: heartbeatInterval,
		RequestTimeout:    requestTimeout,
	})

	if err := agent.Run(ctx); err != nil {
		logs.Errorf("Edge Agent stopped: %v", err)
		return err
	}
	return nil
}
