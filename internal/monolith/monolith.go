package monolith

import (
	"context"
	"net/http"
)

type Monolith interface {
	Mux() *http.ServeMux
}

type Module interface {
	Start(ctx context.Context, mono Monolith) error
}
