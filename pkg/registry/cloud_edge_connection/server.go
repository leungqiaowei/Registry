package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type EdgeNode struct {
	Conn    net.Conn
	ID      string
	Address string
}

type CloudServer struct {
	Listener    net.Listener
	EdgeNodes   sync.Map // 存储所有连接的边节点
	MessageChan chan *Message
}

type Message struct {
	From    string
	To      string
	Content string
	Type    string // "control", "data", "heartbeat"
	Time    time.Time
}

func NewCloudServer(port string) (*CloudServer, error) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, err
	}

	return &CloudServer{
		Listener:    listener,
		MessageChan: make(chan *Message, 100),
	}, nil
}

func (cs *CloudServer) Start() {
	log.Printf("Cloud server started on %s", cs.Listener.Addr())

	go cs.handleMessages()

	for {
		conn, err := cs.Listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}

		go cs.handleConnection(conn)
	}
}

func (cs *CloudServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// 读取边节点ID
	reader := bufio.NewReader(conn)
	nodeID, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Failed to read node ID: %v", err)
		return
	}
	nodeID = nodeID[:len(nodeID)-1] // 去除换行符

	edgeNode := &EdgeNode{
		Conn:    conn,
		ID:      nodeID,
		Address: conn.RemoteAddr().String(),
	}

	cs.EdgeNodes.Store(nodeID, edgeNode)
	log.Printf("Edge node %s connected from %s", nodeID, edgeNode.Address)

	// 发送欢迎消息
	welcomeMsg := fmt.Sprintf("Welcome edge node %s, connection established at %s\n",
		nodeID, time.Now().Format("2006-01-02 15:04:05"))
	conn.Write([]byte(welcomeMsg))

	// 心跳检测和消息处理
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 发送心跳
			_, err := conn.Write([]byte("HEARTBEAT\n"))
			if err != nil {
				log.Printf("Edge node %s disconnected: %v", nodeID, err)
				cs.EdgeNodes.Delete(nodeID)
				return
			}

		default:
			// 设置读取超时
			conn.SetReadDeadline(time.Now().Add(35 * time.Second))

			message, err := reader.ReadString('\n')
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue // 超时继续
				}
				log.Printf("Edge node %s disconnected: %v", nodeID, err)
				cs.EdgeNodes.Delete(nodeID)
				return
			}

			message = message[:len(message)-1] // 去除换行符

			if message == "PONG" {
				log.Printf("Received heartbeat from edge node %s", nodeID)
				continue
			}

			// 处理业务消息
			cs.MessageChan <- &Message{
				From:    nodeID,
				Content: message,
				Type:    "data",
				Time:    time.Now(),
			}

			log.Printf("Received from %s: %s", nodeID, message)

			// 回复确认
			response := fmt.Sprintf("ACK: %s\n", message)
			conn.Write([]byte(response))
		}
	}
}

func (cs *CloudServer) handleMessages() {
	for msg := range cs.MessageChan {
		switch msg.Type {
		case "data":
			log.Printf("Processing message from %s: %s", msg.From, msg.Content)
			// 这里可以添加业务逻辑处理

		case "control":
			log.Printf("Control message from %s: %s", msg.From, msg.Content)

		case "heartbeat":
			// 心跳消息已在上层处理
		}
	}
}

func (cs *CloudServer) SendToEdge(nodeID string, message string) error {
	if value, ok := cs.EdgeNodes.Load(nodeID); ok {
		edgeNode := value.(*EdgeNode)
		_, err := edgeNode.Conn.Write([]byte(message + "\n"))
		return err
	}
	return fmt.Errorf("edge node %s not found", nodeID)
}

func (cs *CloudServer) Broadcast(message string) {
	cs.EdgeNodes.Range(func(key, value interface{}) bool {
		edgeNode := value.(*EdgeNode)
		edgeNode.Conn.Write([]byte(message + "\n"))
		return true
	})
}

func (cs *CloudServer) Stop() {
	cs.Listener.Close()
	close(cs.MessageChan)

	// 关闭所有连接
	cs.EdgeNodes.Range(func(key, value interface{}) bool {
		edgeNode := value.(*EdgeNode)
		edgeNode.Conn.Close()
		return true
	})
}

func main() {
	server, err := NewCloudServer("8119")
	if err != nil {
		log.Fatal(err)
	}

	// 优雅关闭
	defer server.Stop()

	// 启动控制台命令处理
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			cmd := scanner.Text()
			if cmd == "list" {
				fmt.Println("Connected edge nodes:")
				server.EdgeNodes.Range(func(key, value interface{}) bool {
					edgeNode := value.(*EdgeNode)
					fmt.Printf(" - %s (%s)\n", edgeNode.ID, edgeNode.Address)
					return true
				})
			} else if strings.HasPrefix(cmd, "send ") {
				parts := strings.SplitN(cmd[5:], " ", 2)
				if len(parts) == 2 {
					err := server.SendToEdge(parts[0], parts[1])
					if err != nil {
						fmt.Printf("Send error: %v\n", err)
					} else {
						fmt.Println("Message sent")
					}
				}
			} else if strings.HasPrefix(cmd, "broadcast ") {
				server.Broadcast(cmd[10:])
				fmt.Println("Broadcast sent")
			}
		}
	}()

	server.Start()
}
