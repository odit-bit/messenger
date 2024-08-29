package chat

import (
	"context"

	"github.com/odit-bit/messenger/internal/monolith"
)

type Module struct{}

func (mod *Module) Start(ctx context.Context, mono monolith.Monolith) error {
	app := New()
	RegisterHandler(context.TODO(), app, mono.Mux())
	return nil
}
