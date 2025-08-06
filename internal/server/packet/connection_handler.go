package packet

import (
	"bufio"
	logger "github.com/Skylite-Dev-Team/skylite-fsd/internal/config"
	"net"
)

var (
	splitSign    = []byte("\r\n")
	splitSignLen = len(splitSign)
)

type ConnectionHandler struct {
	Conn   net.Conn
	ConnId string
}

func (c *ConnectionHandler) handleLine(line []byte) {
	command, data := parserCommandLine(line)
	logger.DebugF("[%s] %s %v", c.ConnId, command, data)
}

func (c *ConnectionHandler) HandleConnection() {
	defer func() {
		logger.DebugF("[%s] Connection closed", c.ConnId)
		if err := c.Conn.Close(); err != nil && !isNetClosedError(err) {
			logger.WarnF("[%s] Error occurred while closing connection, details: %v", c.ConnId, err)
		}
	}()
	scanner := bufio.NewScanner(c.Conn)
	scanner.Split(createSplitFunc(splitSign))
	for scanner.Scan() {
		line := scanner.Bytes()
		line = line[:len(line)-splitSignLen]
		logger.DebugF("[%s] %s", c.ConnId, line)
		c.handleLine(line)
	}

	if err := scanner.Err(); err != nil {
		logger.InfoF("Read error: %v", err.Error())
	}
}
