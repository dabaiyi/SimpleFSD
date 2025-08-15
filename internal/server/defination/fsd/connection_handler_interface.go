// Package fsd
package fsd

import (
	"github.com/half-nothing/fsd-server/internal/server/database"
	"net"
)

type ConnectionHandlerInterface interface {
	SendError(result *Result)
	HandleConnection()
	Callsign() string
	SetCallsign(callsign string)
	User() *database.User
	SetUser(user *database.User)
	ConnId() string
	Conn() net.Conn
	SetDisconnected(disconnect bool)
}
