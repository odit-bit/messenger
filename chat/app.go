package chat

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type App struct {
	chats map[string]*Chat
}

func New() *App {
	return &App{
		chats: map[string]*Chat{},
	}
}

func (app *App) List() ([]string, error) {
	result := []string{}
	for _, chat := range app.chats {
		result = append(result, chat.Name)
	}

	return result, nil
}

func (app *App) GetChat(name string) (*Chat, bool, error) {
	v, ok := app.chats[name]
	return v, ok, nil
}

func (app *App) CreateChat(name string) (*Chat, error) {
	if _, ok := app.chats[name]; ok {
		return nil, fmt.Errorf("chat name existed")
	}

	c := NewChat(name)
	app.chats[name] = c
	return c, nil
}

func (app *App) JoinChat(name string, ws *websocket.Conn) error {
	chat, ok := app.chats[name]
	if !ok {
		return fmt.Errorf("room not existed")
	}
	p := Participant{
		conn:     ws,
		messages: make(chan []byte),
		chat:     chat,
	}
	chat.AddParticipant(&p)
	app.chats[name] = chat
	return nil
}

func (app *App) IsChatExist(name string) bool {
	if _, ok := app.chats[name]; !ok {
		return false
	}
	return true
}
