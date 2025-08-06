package packet

import (
	"fmt"
	logger "github.com/Skylite-Dev-Team/skylite-fsd/internal/config"
	"time"
)

type Heartbeat func() error

type HeartbeatSender struct {
	interval time.Duration
	ticker   *time.Ticker
	stopChan chan struct{}
	sendFunc Heartbeat
}

func NewHeartbeatSender(interval time.Duration, sendFunc Heartbeat) *HeartbeatSender {
	return &HeartbeatSender{
		interval: interval,
		stopChan: make(chan struct{}),
		sendFunc: sendFunc,
	}
}

func (h *HeartbeatSender) Start() {
	h.ticker = time.NewTicker(h.interval)

	go func() {
		defer fmt.Println("Heartbeat sender stopped")

		for {
			select {
			case <-h.ticker.C:
				err := h.sendFunc()
				if err != nil {
					logger.ErrorF("Error sending heartbeat: %v\n", err)
				}
			case <-h.stopChan:
				return
			}
		}
	}()
}

func (h *HeartbeatSender) Stop() {
	if h.ticker != nil {
		h.ticker.Stop()
	}
	close(h.stopChan)
}
