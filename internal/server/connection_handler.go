package server

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/Skylite-Dev-Team/skylite-fsd/internal/logger"
	"net"
)

// ConnectionHandler 处理FSD客户端连接
type ConnectionHandler struct {
	conn   net.Conn // 网络连接
	connId string   // 连接ID
}

func IsNetClosedError(err error) bool {
	if errors.Is(err, net.ErrClosed) {
		return true
	}
	var opErr *net.OpError
	ok := errors.As(err, &opErr)
	return ok && opErr.Timeout()
}

func createSplitFunc(sep []byte) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// 数据结束且无剩余数据
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		// 查找分隔符位置
		if i := bytes.Index(data, sep); i >= 0 {
			// 返回分隔符前的内容（不包括分隔符）
			return i + len(sep), data[0:i], nil
		}

		// 数据结束但未找到分隔符：返回剩余数据
		if atEOF {
			return len(data), data, nil
		}

		// 请求更多数据
		return 0, nil, nil
	}
}

// handleConnection 处理完整的连接生命周期
func (c *ConnectionHandler) handleConnection() {
	// 确保连接最终被关闭
	defer func() {
		logger.DebugF("[%s] Connection closed", c.connId)
		if err := c.conn.Close(); err != nil && !IsNetClosedError(err) {
			logger.WarnF("[%s] Error occurred while closing connection, details: %v", c.connId, err)
		}
	}()
	scanner := bufio.NewScanner(c.conn)
	scanner.Split(createSplitFunc([]byte("\r\n")))
	for scanner.Scan() {
		line := scanner.Bytes()
		logger.InfoF("Received: %q\n", line)
	}

	if err := scanner.Err(); err != nil {
		logger.InfoF("Read error: %v", err.Error())
	}
}
