// Package fsd
package fsd

import (
	"github.com/half-nothing/fsd-server/internal/interfaces/operation"
	"net"
)

type ConnectionHandlerInterface interface {
	SendError(result *Result)
	HandleConnection()
	Callsign() string
	SetCallsign(callsign string)
	User() *operation.User
	SetUser(user *operation.User)
	ConnId() string
	Conn() net.Conn
	SetDisconnected(disconnect bool)
}
