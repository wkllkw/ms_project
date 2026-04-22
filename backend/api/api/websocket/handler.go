package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 毕设项目允许跨域
	},
}

// Client 表示一个已连接的客户端
type Client struct {
	ID         string          // 客户端唯一ID (client_id)
	UserID     string          // 绑定的用户ID
	Conn       *websocket.Conn // WebSocket连接
	Send       chan []byte     // 发送消息通道
	LastActive time.Time       // 最后活跃时间
	mu         sync.Mutex      // 保护 UserID 写入
}

// Message 表示 WebSocket 通信的消息格式（与前端 socket.vue 对齐）
type Message struct {
	Action string      `json:"action"` // connect / ping / notify / bind_user
	Data   interface{} `json:"data"`
}

// ClientMap 管理所有在线客户端
type ClientMap struct {
	clients    map[string]*Client // client_id -> Client
	userClient map[string]string  // user_id -> client_id
	mu         sync.RWMutex
}

// Manager WebSocket 全局管理器
var Manager = &ClientMap{
	clients:    make(map[string]*Client),
	userClient: make(map[string]string),
}

// Register 注册一个新连接
func (m *ClientMap) Register(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[client.ID] = client
}

// Unregister 移除一个连接
func (m *ClientMap) Unregister(clientID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if client, ok := m.clients[clientID]; ok {
		if client.UserID != "" {
			delete(m.userClient, client.UserID)
		}
		close(client.Send)
		delete(m.clients, clientID)
	}
}

// BindUser 绑定用户ID到客户端（对应前端 bindClientId）
func (m *ClientMap) BindUser(clientID, userID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	client, ok := m.clients[clientID]
	if !ok {
		return false
	}
	client.UserID = userID
	m.userClient[userID] = clientID
	return true
}

// OnlineCount 返回在线人数
func (m *ClientMap) OnlineCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients)
}

// Broadcast 向所有在线用户广播消息
func (m *ClientMap) Broadcast(message Message) {
	data, _ := json.Marshal(message)
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, client := range m.clients {
		select {
		case client.Send <- data:
		default:
			// 发送缓冲满，跳过该客户端
		}
	}
}

// SendToUser 向指定用户发送消息
func (m *ClientMap) SendToUser(userID string, message Message) {
	data, _ := json.Marshal(message)
	m.mu.RLock()
	clientID, ok := m.userClient[userID]
	m.mu.RUnlock()
	if !ok {
		return
	}
	m.mu.RLock()
	client, ok := m.clients[clientID]
	m.mu.RUnlock()
	if ok {
		select {
		case client.Send <- data:
		default:
		}
	}
}

// HandleWebSocket 处理 WebSocket 连接请求
func HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Printf("websocket upgrade error: %v\n", err)
		return
	}

	// 生成唯一的 client_id
	clientID := uuid.New().String()

	client := &Client{
		ID:         clientID,
		Conn:       conn,
		Send:       make(chan []byte, 256),
		LastActive: time.Now(),
	}

	Manager.Register(client)

	// 启动读写协程
	go client.writePump()
	go client.readPump()

	// 发送 connect 消息给前端（触发前端的 bindClientId 流程）
	connectMsg := Message{
		Action: "connect",
		Data: map[string]interface{}{
			"client_id": clientID,
			"online":    Manager.OnlineCount(),
		},
	}
	connectData, _ := json.Marshal(connectMsg)
	client.Send <- connectData
}

// readPump 读取客户端消息
func (c *Client) readPump() {
	defer func() {
		Manager.Unregister(c.ID)
		c.Conn.Close()
	}()

	// 设置读超时（心跳检测：90秒无数据则断开）
	c.Conn.SetReadLimit(65536)           // 最大消息 64KB
	c.Conn.SetReadDeadline(time.Now().Add(90 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.LastActive = time.Now()
		c.Conn.SetReadDeadline(time.Now().Add(90 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg.Action {
		case "ping":
			// 前端发送 ping，回复 pong
			c.LastActive = time.Now()
			pongData, _ := json.Marshal(Message{Action: "pong", Data: gin.H{"time": time.Now().Unix()}})
			c.Send <- pongData

		case "bind_user":
			// 前端调用 bindClientId 后绑定 user_id
			if uid, ok := msg.Data.(map[string]interface{})["uid"].(string); ok && uid != "" {
				Manager.BindUser(c.ID, uid)
			}
		}
	}
}

// writePump 向客户端写入消息
func (c *Client) writePump() {
	// 定时发送 ping（30秒一次），确保连接保活
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 批量写入队列中待发消息
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.Send)
			}
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
