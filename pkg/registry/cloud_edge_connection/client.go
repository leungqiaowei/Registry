package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

type EdgeClient struct {
	Conn        net.Conn
	CloudAddr   string
	NodeID      string
	Reconnect   bool
	MessageChan chan string
	StopChan    chan bool
}

func NewEdgeClient(cloudAddr, nodeID string) *EdgeClient {
	return &EdgeClient{
		CloudAddr:   cloudAddr,
		NodeID:      nodeID,
		Reconnect:   true,
		MessageChan: make(chan string, 100),
		StopChan:    make(chan bool),
	}
}

func (ec *EdgeClient) Connect() error {
	conn, err := net.Dial("tcp", ec.CloudAddr)
	if err != nil {
		return err
	}

	ec.Conn = conn

	// 发送节点ID作为身份标识
	_, err = conn.Write([]byte(ec.NodeID + "\n"))
	if err != nil {
		conn.Close()
		return err
	}

	log.Printf("Connected to cloud server %s", ec.CloudAddr)
	return nil
}

func (ec *EdgeClient) Start() {
	defer ec.Conn.Close()

	reader := bufio.NewReader(ec.Conn)

	// 读取欢迎消息
	welcome, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Failed to read welcome message: %v", err)
		return
	}
	log.Printf("Cloud server: %s", strings.TrimSpace(welcome))

	// 启动消息发送协程
	go ec.handleSend()

	// 主循环处理接收消息
	for {
		select {
		case <-ec.StopChan:
			log.Println("Stopping edge client")
			return

		default:
			// 设置读取超时
			ec.Conn.SetReadDeadline(time.Now().Add(35 * time.Second))

			message, err := reader.ReadString('\n')
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue // 超时继续
				}
				log.Printf("Connection lost: %v", err)
				ec.reconnect()
				return
			}

			message = strings.TrimSpace(message)
			ec.handleMessage(message)
		}
	}
}

func (ec *EdgeClient) handleMessage(message string) {
	switch message {
	case "HEARTBEAT":
		// 响应心跳
		ec.Conn.Write([]byte("PONG\n"))
		log.Printf("Heartbeat responded")

	default:
		if strings.HasPrefix(message, "ACK:") {
			log.Printf("Cloud acknowledged: %s", message[4:])
		} else {
			log.Printf("Received from cloud: %s", message)
			// 这里可以添加业务逻辑处理云端的消息
		}
	}
}

func (ec *EdgeClient) handleSend() {
	ticker := time.NewTicker(60 * time.Second) // 每60秒发送一次数据
	defer ticker.Stop()

	counter := 0
	for {
		select {
		case <-ec.StopChan:
			return

		case msg := <-ec.MessageChan:
			// 发送用户输入的消息
			_, err := ec.Conn.Write([]byte(msg + "\n"))
			if err != nil {
				log.Printf("Failed to send message: %v", err)
			}

		case <-ticker.C:
			// 定时发送示例数据
			counter++
			data := fmt.Sprintf("Edge data #%d at %s", counter, time.Now().Format("15:04:05"))
			_, err := ec.Conn.Write([]byte(data + "\n"))
			if err != nil {
				log.Printf("Failed to send data: %v", err)
			} else {
				log.Printf("Sent: %s", data)
			}
		}
	}
}

func (ec *EdgeClient) SendMessage(message string) {
	ec.MessageChan <- message
}

func (ec *EdgeClient) Stop() {
	ec.Reconnect = false
	close(ec.StopChan)
	if ec.Conn != nil {
		ec.Conn.Close()
	}
}

func (ec *EdgeClient) reconnect() {
	if !ec.Reconnect {
		return
	}

	log.Println("Attempting to reconnect...")

	for {
		time.Sleep(5 * time.Second) // 5秒后重试

		err := ec.Connect()
		if err != nil {
			log.Printf("Reconnect failed: %v, retrying...", err)
			continue
		}

		log.Println("Reconnected successfully")
		go ec.Start()
		break
	}
}

func main() {

	cloudAddr := "223.166.61.57:11006"
	nodeID := "test"

	client := NewEdgeClient(cloudAddr, nodeID)

	// 首次连接
	err := client.Connect()
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// 启动控制台输入处理
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := scanner.Text()
			if text == "quit" {
				client.Stop()
				os.Exit(0)
			} else if text == "status" {
				log.Printf("Edge client status: connected to %s", cloudAddr)
			} else if strings.TrimSpace(text) != "" {
				client.SendMessage(text)
			}
		}
	}()

	// 主循环
	client.Start()

	// 如果连接断开，开始重连
	client.reconnect()
}
