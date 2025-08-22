// Package interfaces
package interfaces

import (
	c "github.com/half-nothing/fsd-server/internal/config"
	"github.com/half-nothing/fsd-server/internal/interfaces/operation"
)

type ApplicationContent struct {
	config *c.Config
	*operation.DatabaseOperations
}

func NewApplicationContent(
	config *c.Config,
	db *operation.DatabaseOperations,
) *ApplicationContent {
	return &ApplicationContent{config, db}
}

func (app *ApplicationContent) Config() *c.Config {
	return app.config
}
