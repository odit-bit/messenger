package chat

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Handler struct {
	app *App
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (h *Handler) handleCreateChat(w http.ResponseWriter, r *http.Request) {
	// upgrage connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		conn.Close()
		return
	}

	name := r.PathValue("chat_name")
	if h.app.IsChatExist(name) {
		wsChatExistError(conn)
		conn.Close()
		return
	}

	chat, err := h.app.CreateChat(name)
	if err != nil {
		wsError(conn, err)
		conn.Close()
		return
	}

	if err := h.app.JoinChat(name, conn); err != nil {
		wsError(conn, err)
		conn.Close()
		return
	}
	go chat.Run()
}

func (h *Handler) handleJoinChat(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("chat_name")

	// upgrage connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Println(err)
		conn.Close()
		return
	}

	if err := h.app.JoinChat(name, conn); err != nil {
		// http.Error(w, err.Error(), 500)
		conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		conn.Close()
		return
	} else {
		conn.WriteMessage(websocket.TextMessage, []byte("ok"))
	}
}

func (h *Handler) handleChatList(w http.ResponseWriter, r *http.Request) {

	result, _ := h.app.List()
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Println("failed encode chats: $v", err)
		http.Error(w, err.Error(), 500)
		return
	}
}

func (h *Handler) handleChatMessages(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("chat_name")
	chat, ok, _ := h.app.GetChat(name)
	if !ok {
		http.Error(w, "chat not found", http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(chat.Messages()); err != nil {
		log.Println(err)
		return
	}

}

func RegisterHandler(ctx context.Context, app *App, mux *http.ServeMux) {
	h := Handler{
		app: app,
	}
	//websocket upgrade
	mux.HandleFunc("/chats/create/{chat_name}", h.handleCreateChat)
	mux.HandleFunc("/chats/join/{chat_name}", h.handleJoinChat)

	//rest-http
	mux.HandleFunc("/chats", h.handleChatList)
	mux.HandleFunc("/chats/messages/{chat_name}", h.handleChatMessages)
}

// ERROR

func wsChatExistError(conn *websocket.Conn) {
	conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(
			websocket.ClosePolicyViolation,
			"chat is exist",
		))
}

func wsError(conn *websocket.Conn, err error) {
	conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(
			websocket.ClosePolicyViolation,
			err.Error(),
		))
}
