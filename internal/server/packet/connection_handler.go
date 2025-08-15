package packet

import (
	"bufio"
	"fmt"
	logger "github.com/half-nothing/fsd-server/internal/config"
	"github.com/half-nothing/fsd-server/internal/server/database"
	. "github.com/half-nothing/fsd-server/internal/server/defination/fsd"
	"net"
	"sync/atomic"
	"time"
)

var (
	splitSign    = []byte("\r\n")
	splitSignLen = len(splitSign)
)

type ConnectionHandler struct {
	conn         net.Conn
	connId       string
	callsign     string
	client       ClientInterface
	user         *database.User
	disconnected atomic.Bool
}

func NewConnectionHandler(conn net.Conn, connId string) *ConnectionHandler {
	return &ConnectionHandler{
		conn:         conn,
		connId:       connId,
		callsign:     "unknown",
		client:       nil,
		user:         nil,
		disconnected: atomic.Bool{},
	}
}

func (c *ConnectionHandler) SendError(result *Result) {
	if result.Success {
		return
	}
	if c.client != nil {
		c.client.SendError(result)
		return
	}
	packet := makePacket(Error, "server", c.callsign, fmt.Sprintf("%03d", result.Errno.Index()), result.Env, result.Errno.String())
	logger.DebugF("[%s](%s) <- %s", c.connId, c.callsign, packet[:len(packet)-splitSignLen])
	_, _ = c.conn.Write(packet)
	if result.Fatal {
		c.disconnected.Store(true)
		time.AfterFunc(500*time.Millisecond, func() {
			_ = c.conn.Close()
		})
	}
}

func (c *ConnectionHandler) handleLine(line []byte) {
	if c.disconnected.Load() {
		return
	}
	command, data := parserCommandLine(line)
	result := c.handleCommand(command, data, line)
	if result == nil {
		logger.WarnF("[%s](%s) handleCommand return a nil result", c.connId, c.callsign)
		return
	}
	if !result.Success {
		logger.ErrorF("[%s](%s) handleCommand fail, %s, %s", c.connId, c.callsign, result.Errno.String(), result.Err.Error())
		c.SendError(result)
	}
}

func (c *ConnectionHandler) HandleConnection() {
	defer func() {
		logger.DebugF("[%s](%s) x Connection closed", c.connId, c.callsign)
		if err := c.conn.Close(); err != nil && !isNetClosedError(err) {
			logger.WarnF("[%s](%s) Error occurred while closing connection, details: %v", c.connId, c.callsign, err)
		}
	}()
	scanner := bufio.NewScanner(c.conn)
	scanner.Split(createSplitFunc(splitSign))
	for scanner.Scan() {
		line := scanner.Bytes()
		logger.DebugF("[%s](%s) -> %s", c.connId, c.callsign, line)
		c.handleLine(line)
		if c.disconnected.Load() {
			break
		}
	}

	if c.client != nil {
		if c.client.IsAtc() {
			clientManager.BroadcastMessage(makePacket(RemoveAtc, c.client.Callsign(), "SERVER"), c.client, BroadcastToClientInRange)
		} else {
			clientManager.BroadcastMessage(makePacket(RemovePilot, c.client.Callsign(), "SERVER"), c.client, BroadcastToClientInRange)
		}
		c.client.MarkedDisconnect(false)
	}
}

func (c *ConnectionHandler) Callsign() string { return c.callsign }

func (c *ConnectionHandler) SetCallsign(callsign string) { c.callsign = callsign }

func (c *ConnectionHandler) User() *database.User { return c.user }

func (c *ConnectionHandler) SetUser(user *database.User) { c.user = user }

func (c *ConnectionHandler) ConnId() string { return c.connId }

func (c *ConnectionHandler) Conn() net.Conn { return c.conn }

func (c *ConnectionHandler) SetDisconnected(disconnect bool) { c.disconnected.Store(disconnect) }
