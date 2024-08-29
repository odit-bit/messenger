package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/odit-bit/messenger/chat"
)

var chats map[string]*chat.Chat = map[string]*chat.Chat{}

func main() {

	mux := http.NewServeMux()

	// mux.HandleFunc("/ws", serveWS)
	// mux.HandleFunc("/room", roomChat)
	// mux.HandleFunc("/join", joinChat)
	mux.HandleFunc("GET /chats", handleListChat)
	mux.HandleFunc("GET /chats/messages/{chat_name}", handleChatMessages)
	err := http.ListenAndServe(":8989", mux)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleListChat(w http.ResponseWriter, r *http.Request) {
	result := []string{}
	for _, chat := range chats {
		result = append(result, chat.Name)
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Println("failed encode chats: $v", err)
		http.Error(w, err.Error(), 500)
		return
	}
}

func handleChatMessages(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("chat_name")
	chat, ok := chats[name]
	if !ok {
		http.Error(w, "chat not found", http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(chat.Messages()); err != nil {
		log.Println(err)
		return
	}

}
