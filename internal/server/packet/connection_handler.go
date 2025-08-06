package packet

import (
	"bufio"
	"fmt"
	logger "github.com/Skylite-Dev-Team/skylite-fsd/internal/config"
	"github.com/Skylite-Dev-Team/skylite-fsd/internal/server/database"
	"net"
	"time"
)

var (
	splitSign    = []byte("\r\n")
	splitSignLen = len(splitSign)
)

type ConnectionHandler struct {
	Conn   net.Conn
	ConnId string
	Client *Client
	User   *database.User
}

func (c *ConnectionHandler) SendLine(line []byte) {
	logger.DebugF("[%s] -> %s", c.ConnId, line)
	_, _ = c.Conn.Write(line)
}

func (c *ConnectionHandler) SendError(result *Result) {
	if result.success {
		return
	}
	var packet []byte
	if c.Client != nil {
		packet = makePacket(Error, "server", c.Client.Callsign, fmt.Sprintf("%03d", result.errno.Index()), result.env, result.errno.String())
	} else {
		packet = makePacket(Error, "server", "unknown", fmt.Sprintf("%03d", result.errno.Index()), result.env, result.errno.String())
	}
	c.SendLine(packet)
	if result.fatal {
		time.AfterFunc(100*time.Millisecond, func() {
			_ = c.Conn.Close()
		})
	}
}

func (c *ConnectionHandler) handleLine(line []byte) {
	command, data := parserCommandLine(line)
	result := c.HandleCommand(command, data, line)
	if !result.success {
		c.SendError(result)
	}
}

func (c *ConnectionHandler) HandleConnection() {
	defer func() {
		logger.DebugF("[%s] x Connection closed", c.ConnId)
		if err := c.Conn.Close(); err != nil && !isNetClosedError(err) {
			logger.WarnF("[%s] Error occurred while closing connection, details: %v", c.ConnId, err)
		}
	}()
	scanner := bufio.NewScanner(c.Conn)
	scanner.Split(createSplitFunc(splitSign))
	for scanner.Scan() {
		line := scanner.Bytes()
		logger.DebugF("[%s] <- %s", c.ConnId, line)
		c.handleLine(line)
	}

	if err := scanner.Err(); err != nil {
		logger.InfoF("Read error: %v", err.Error())
	}

	if c.Client != nil {
		logger.DebugF("[%s] Disconnected, session will be saved for %s", c.Client.Callsign, config.SessionCleanDuration.String())
		c.Client.MarkedDisconnect()
	}
}
