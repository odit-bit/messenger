package main

import (
	"context"
	"log"
	"net/http"

	"github.com/odit-bit/messenger/chat"
	"github.com/odit-bit/messenger/internal/monolith"
)

func main() {
	root := Root{
		modules: []monolith.Module{&chat.Module{}},
		mux:     http.NewServeMux(),
	}

	if err := root.Run(context.TODO()); err != nil {
		log.Fatal(err)
	}

}

type Root struct {
	modules []monolith.Module
	mux     *http.ServeMux
}

func (s *Root) Mux() *http.ServeMux {
	return s.mux
}

func (s *Root) Run(ctx context.Context) error {
	for _, mod := range s.modules {
		if err := mod.Start(ctx, s); err != nil {
			return err
		}
	}

	err := http.ListenAndServe(":8989", s.mux)
	if err != nil {
		return err
	}
	return nil
}
