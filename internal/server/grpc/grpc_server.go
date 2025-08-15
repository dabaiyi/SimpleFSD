package __

import (
	"github.com/half-nothing/fsd-server/internal/server/packet"
	"golang.org/x/net/context"
	"sync"
	"time"
)

type GrpcServer struct {
	generateTime time.Time
	onlineClient *OnlineClient
	mu           sync.RWMutex
	queryCache   time.Duration
}

func NewGrpcServer(queryCache time.Duration) *GrpcServer {
	return &GrpcServer{
		generateTime: time.Now(),
		onlineClient: nil,
		mu:           sync.RWMutex{},
		queryCache:   queryCache,
	}
}

func (s *GrpcServer) GetOnlineClient(_ context.Context, _ *Empty) (*OnlineClient, error) {
	s.mu.RLock()
	if s.onlineClient != nil && time.Since(s.generateTime) <= s.queryCache {
		defer s.mu.RUnlock()
		return s.onlineClient, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.onlineClient != nil && time.Since(s.generateTime) <= s.queryCache {
		return s.onlineClient, nil
	}

	clientManager := packet.GetClientManager()
	clientCopy := clientManager.GetClientSnapshot()
	defer clientManager.PutSlice(clientCopy)

	s.onlineClient = &OnlineClient{
		TotalOnline: 0,
		PilotOnline: 0,
		AtcOnline:   0,
		OnlineAtc:   make([]*OnlineAtc, 0),
		OnlinePilot: make([]*OnlinePilot, 0),
	}

	for _, client := range clientCopy {
		if client == nil || client.Disconnected() {
			continue
		}
		s.onlineClient.TotalOnline++
		if client.IsAtc() {
			s.onlineClient.AtcOnline++
			atcInfo := &OnlineAtc{
				Callsign:   client.Callsign(),
				Username:   client.User().Username,
				Email:      client.User().Email,
				Cid:        int32(client.User().Cid),
				RealName:   client.RealName(),
				Lat:        float32(client.Position()[0].Latitude),
				Lon:        float32(client.Position()[0].Longitude),
				Rating:     int32(client.Rating().Index()),
				Facility:   client.Facility().String(),
				Frequency:  int32(client.Frequency() + 100000),
				AtcInfo:    client.AtisInfo(),
				OnlineTime: client.History().StartTime.Unix(),
			}
			s.onlineClient.OnlineAtc = append(s.onlineClient.OnlineAtc, atcInfo)
		} else {
			s.onlineClient.PilotOnline++
			pilotInfo := &OnlinePilot{
				Callsign:    client.Callsign(),
				Username:    client.User().Username,
				Email:       client.User().Email,
				Cid:         int32(client.User().Cid),
				RealName:    client.RealName(),
				Lat:         float32(client.Position()[0].Latitude),
				Lon:         float32(client.Position()[0].Longitude),
				Altitude:    int32(client.Altitude()),
				GroundSpeed: int32(client.GroundSpeed()),
				OnlineTime:  client.History().StartTime.Unix(),
			}
			s.onlineClient.OnlinePilot = append(s.onlineClient.OnlinePilot, pilotInfo)
		}
	}
	s.generateTime = time.Now()

	return s.onlineClient, nil
}

func (s *GrpcServer) mustEmbedUnimplementedServerStatusServer() {
}
