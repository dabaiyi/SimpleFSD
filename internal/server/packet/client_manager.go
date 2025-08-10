package packet

import (
	"context"
	"fmt"
	c "github.com/half-nothing/fsd-server/internal/config"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type ClientManager struct {
	clients         map[string]*Client
	lock            sync.RWMutex
	shuttingDown    atomic.Bool
	clientSlicePool sync.Pool
}

var (
	clientManager   *ClientManager
	heartbeatSender *HeartbeatSender
	config          *c.Config
	once            sync.Once
)

func GetClientManager() *ClientManager {
	once.Do(func() {
		config, _ = c.GetConfig()
		if clientManager == nil {
			clientManager = &ClientManager{
				clients:      make(map[string]*Client),
				shuttingDown: atomic.Bool{},
				clientSlicePool: sync.Pool{
					New: func() interface{} {
						return make([]*Client, 0, 128)
					},
				},
			}
			heartbeatSender = NewHeartbeatSender(config.Server.FSDServer.HeartbeatDuration, clientManager.SendHeartBeat)
		}
	})
	return clientManager
}

func (cm *ClientManager) PutSlice(clients []*Client) {
	cm.clientSlicePool.Put(clients)
}

func (cm *ClientManager) Shutdown(ctx context.Context) error {
	if !cm.shuttingDown.CompareAndSwap(false, true) {
		return fmt.Errorf("shutting down already in progress")
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	heartbeatSender.Stop()

	clients := cm.GetClientSnapshot()
	defer cm.PutSlice(clients)

	done := make(chan struct{})
	go func() {
		defer close(done)
		cm.disconnectClients(clients)
	}()

	select {
	case <-done:
		return nil
	case <-timeoutCtx.Done():
		return timeoutCtx.Err()
	}
}

func (cm *ClientManager) GetClientSnapshot() []*Client {
	cm.lock.RLock()
	defer cm.lock.RUnlock()

	// 从池中获取切片
	clients := cm.clientSlicePool.Get().([]*Client)
	clients = clients[:0]

	// 填充客户端
	for _, client := range cm.clients {
		clients = append(clients, client)
	}
	return clients
}

// 并发断开所有客户端连接
func (cm *ClientManager) disconnectClients(clients []*Client) {
	if len(clients) == 0 {
		return
	}

	sem := make(chan struct{}, config.Server.FSDServer.MaxBroadcastWorkers)
	var wg sync.WaitGroup

	for _, client := range clients {
		wg.Add(1)
		sem <- struct{}{}

		go func(c *Client) {
			defer func() {
				<-sem
				wg.Done()
			}()

			c.MarkedDisconnect(true)
		}(client)
	}

	wg.Wait()
}

func (cm *ClientManager) SendHeartBeat() error {
	if cm.shuttingDown.Load() {
		return nil
	}
	randomInt := rand.Int()
	packet := makePacket(WindDelta, "SERVER", string(AllClient), strconv.Itoa(randomInt%11-5), strconv.Itoa(randomInt%21-10))
	cm.BroadcastMessage(packet, nil, BroadcastToAll)
	return nil
}

func (cm *ClientManager) AddClient(client *Client) error {
	if cm.shuttingDown.Load() {
		return fmt.Errorf("server shutting down")
	}
	cm.lock.Lock()
	defer cm.lock.Unlock()

	if _, exists := cm.clients[client.Callsign]; exists {
		return fmt.Errorf("client already registered: %s", client.Callsign)
	}
	cm.clients[client.Callsign] = client
	return nil
}

func (cm *ClientManager) GetClient(callsign string) (*Client, bool) {
	if cm.shuttingDown.Load() {
		return nil, false
	}

	cm.lock.RLock()
	defer cm.lock.RUnlock()

	client, exists := cm.clients[callsign]
	return client, exists
}

func (cm *ClientManager) DeleteClient(callsign string) bool {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	if _, exists := cm.clients[callsign]; !exists {
		return false
	}

	delete(cm.clients, callsign)
	return true
}

func (cm *ClientManager) SendMessageTo(callsign string, message []byte) error {
	if cm.shuttingDown.Load() {
		return fmt.Errorf("server is shutting down")
	}

	client, exists := cm.GetClient(callsign)
	if !exists {
		return fmt.Errorf("client not found: %s", callsign)
	}

	client.SendLine(message)
	return nil
}

func (cm *ClientManager) BroadcastMessage(message []byte, fromClient *Client, filter BroadcastFilter) {
	if cm.shuttingDown.Load() || len(message) == 0 {
		return
	}

	clients := cm.GetClientSnapshot()
	defer cm.PutSlice(clients) // 重置并放回池中

	if len(clients) == 0 {
		return
	}

	// 准备完整消息（包含分割符）
	fullMsg := make([]byte, len(message), len(message)+len(splitSign))
	copy(fullMsg, message)
	fullMsg = append(fullMsg, splitSign...)

	// 并发广播
	var wg sync.WaitGroup
	sem := make(chan struct{}, config.Server.FSDServer.MaxBroadcastWorkers)

	for _, client := range clients {
		if client == fromClient || client.Disconnected() {
			continue
		}

		if filter != nil && !filter(client, fromClient) {
			continue
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(cl *Client) {
			defer func() {
				<-sem
				wg.Done()
			}()

			c.DebugF("[Broadcast] -> [%s] %s", cl.Callsign, message)
			cl.SendLine(fullMsg)
		}(client)
	}

	wg.Wait()
}
