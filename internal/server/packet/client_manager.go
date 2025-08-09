package packet

import (
	"context"
	"fmt"
	c "github.com/half-nothing/fsd-server/internal/config"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type ClientManager struct {
	clients map[string]*Client
	lock    sync.Mutex
}

type ServerCloseCallback struct {
}

func NewServerCloseCallback() *ServerCloseCallback {
	return &ServerCloseCallback{}
}

func (dc *ServerCloseCallback) Invoke(ctx context.Context) error {
	heartbeatSender.Stop()
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		for _, client := range clientManager.clients {
			client.MarkedDisconnect(true)
		}
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-timeoutCtx.Done():
		return timeoutCtx.Err()
	}
}

var (
	clientManager   *ClientManager
	heartbeatSender *HeartbeatSender
	config          *c.Config
	once            sync.Once
	clientSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]*Client, 0, 128)
		},
	}
)

func GetClientManager() *ClientManager {
	once.Do(func() {
		config, _ = c.GetConfig()
		if clientManager == nil {
			clientManager = &ClientManager{
				clients: make(map[string]*Client),
			}
			heartbeatSender = NewHeartbeatSender(config.Server.FSDServer.HeartbeatDuration, clientManager.SendHeartBeat)
			c.NewCleaner().Add(NewServerCloseCallback())
		}
	})
	return clientManager
}

func (cm *ClientManager) SendHeartBeat() error {
	randomInt := rand.Int()
	packet := makePacket(WindDelta, "SERVER", string(AllClient), strconv.Itoa(randomInt%11-5), strconv.Itoa(randomInt%21-10))
	cm.BroadcastMessage(packet, nil, BroadcastToAll)
	return nil
}

func (cm *ClientManager) AddClient(client *Client) error {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	if _, exists := cm.clients[client.Callsign]; exists {
		return fmt.Errorf("client already registered: %s", client.Callsign)
	}
	cm.clients[client.Callsign] = client
	return nil
}

func (cm *ClientManager) GetClient(callsign string) (*Client, error) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	client, exists := cm.clients[callsign]
	if !exists {
		return nil, fmt.Errorf("client not found: %s", callsign)
	}
	return client, nil
}

func (cm *ClientManager) DeleteClient(callsign string) error {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	if _, exists := cm.clients[callsign]; !exists {
		return fmt.Errorf("client not found: %s", callsign)
	}
	delete(cm.clients, callsign)
	return nil
}

func (cm *ClientManager) SendMessageTo(callsign string, message []byte) error {
	client, err := cm.GetClient(callsign)
	if err != nil {
		return err
	}
	client.SendLine(message)
	return err
}

func (cm *ClientManager) GetClientCopy() []*Client {
	cm.lock.Lock()
	clients := make([]*Client, 0, len(cm.clients))
	for _, client := range cm.clients {
		clients = append(clients, client)
	}
	cm.lock.Unlock()
	return clients
}

func (cm *ClientManager) BroadcastMessage(message []byte, fromClient *Client, filter BroadcastFilter) {
	fullMsg := make([]byte, len(message)+len(splitSign))
	copy(fullMsg, message)
	copy(fullMsg[len(message):], splitSign)

	cm.lock.Lock()

	clients := clientSlicePool.Get().([]*Client)
	clients = clients[:0]

	for _, client := range cm.clients {
		clients = append(clients, client)
	}
	cm.lock.Unlock()

	defer clientSlicePool.Put(clients)

	var wg sync.WaitGroup
	sem := make(chan struct{}, config.Server.FSDServer.MaxBroadcastWorkers)

	for _, client := range clients {
		if client == fromClient || client.disconnect {
			continue
		}
		if !filter(client, fromClient) {
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
			cl.SendLine(message)
		}(client)
	}
	wg.Wait()
}
