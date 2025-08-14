package packet

import (
	"bufio"
	"fmt"
	logger "github.com/half-nothing/fsd-server/internal/config"
	"github.com/half-nothing/fsd-server/internal/server/database"
	"net"
	"sync/atomic"
	"time"
)

var (
	splitSign    = []byte("\r\n")
	splitSignLen = len(splitSign)
)

type ConnectionHandler struct {
	Conn         net.Conn
	ConnId       string
	Callsign     string
	Client       *Client
	User         *database.User
	Disconnected atomic.Bool
}

func (c *ConnectionHandler) SendError(result *Result) {
	if result.success {
		return
	}
	if c.Client != nil {
		c.Client.SendError(result)
		return
	}
	packet := makePacket(Error, "server", c.Callsign, fmt.Sprintf("%03d", result.errno.Index()), result.env, result.errno.String())
	logger.DebugF("[%s](%s) <- %s", c.ConnId, c.Callsign, packet[:len(packet)-splitSignLen])
	_, _ = c.Conn.Write(packet)
	if result.fatal {
		c.Disconnected.Store(true)
		time.AfterFunc(500*time.Millisecond, func() {
			_ = c.Conn.Close()
		})
	}
}

func (c *ConnectionHandler) handleLine(line []byte) {
	if c.Disconnected.Load() {
		return
	}
	command, data := parserCommandLine(line)
	result := c.handleCommand(command, data, line)
	if result == nil {
		logger.WarnF("[%s](%s) handleCommand return a nil result", c.ConnId, c.Callsign)
		return
	}
	if !result.success {
		logger.ErrorF("[%s](%s) handleCommand fail, %s, %s", c.ConnId, c.Callsign, result.errno.String(), result.err.Error())
		c.SendError(result)
	}
}

func (c *ConnectionHandler) HandleConnection() {
	defer func() {
		logger.DebugF("[%s](%s) x Connection closed", c.ConnId, c.Callsign)
		if err := c.Conn.Close(); err != nil && !isNetClosedError(err) {
			logger.WarnF("[%s](%s) Error occurred while closing connection, details: %v", c.ConnId, c.Callsign, err)
		}
	}()
	scanner := bufio.NewScanner(c.Conn)
	scanner.Split(createSplitFunc(splitSign))
	for scanner.Scan() {
		line := scanner.Bytes()
		logger.DebugF("[%s](%s) -> %s", c.ConnId, c.Callsign, line)
		c.handleLine(line)
		if c.Disconnected.Load() {
			break
		}
	}

	if c.Client != nil {
		if c.Client.IsAtc {
			clientManager.BroadcastMessage(makePacket(RemoveAtc, c.Client.Callsign, "SERVER"), c.Client, BroadcastToClientInRange)
		} else {
			clientManager.BroadcastMessage(makePacket(RemovePilot, c.Client.Callsign, "SERVER"), c.Client, BroadcastToClientInRange)
		}
		c.Client.MarkedDisconnect(false)
	}
}
