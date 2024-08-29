package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/odit-bit/messenger/chat"
)

func main() {
	var sender string
	flag.StringVar(&sender, "sender", "", "message sender")
	flag.Parse()

	arg := os.Args[1]
	switch arg {
	case "join":
		if len(os.Args) < 1 {
			log.Fatal("need room name argument")
		}
		chatName := os.Args[2]
		join(sender, chatName)
	case "room":
		room(sender)
	case "chat-ls":
		listChat()

	case "messages":
		if len(os.Args) < 1 {
			log.Fatal("need chat name argument")
		}
		chatName := os.Args[2]
		listMessage(chatName)
	default:
		log.Fatal("unknown arg")
	}
}
func listChat() {

	res, err := http.DefaultClient.Get("http://localhost:8989/chats")
	if err != nil {
		log.Printf("error: %v \n", err)
		return
	}
	defer res.Body.Close()
	chats := []string{}
	if err := json.NewDecoder(res.Body).Decode(&chats); err != nil {
		log.Println(err)
		return
	}
	fmt.Println(chats)
}

func listMessage(chatName string) {

	res, err := http.DefaultClient.Get("http://localhost:8989/chats/messages/" + chatName)
	if err != nil {
		log.Printf("error: %v \n", err)
		return
	}
	defer res.Body.Close()
	messages := []chat.Message{}
	if err := json.NewDecoder(res.Body).Decode(&messages); err != nil {
		log.Println(err)
		return
	}
	for _, msg := range messages {
		fmt.Printf("-- %v \n", msg)
	}

}

func room(sender string) {

	u := url.URL{Scheme: "ws", Host: ":8989", Path: "/room"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Printf("failed dial server: %v", err)
		return
	}
	defer conn.Close()

	if sender == "" {
		sender = conn.LocalAddr().String()
	}

	//handshake or create room
	if err := conn.WriteMessage(websocket.TextMessage, []byte("room-ulala")); err != nil {
		log.Printf("failed create room; %v \n", err)
		return
	}

	// event LOOP here !!!
	HandleMessage(sender, conn)
}

func join(sender, room string) {

	u := url.URL{Scheme: "ws", Host: ":8989", Path: "/join"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Printf("failed dial server: %v", err)
		return
	}
	defer conn.Close()

	if sender == "" {
		sender = conn.LocalAddr().String()
	}

	if err := conn.WriteMessage(websocket.TextMessage, []byte(room)); err != nil {
		log.Println("err")
		return
	}

	if _, msg, err := conn.ReadMessage(); err != nil {
		log.Println(err)
		return
	} else {
		if string(msg) != "ok" {
			log.Println(string(msg))
			conn.Close()
			return
		}
	}

	// event LOOP here !!!
	HandleMessage(sender, conn)
}

type Message struct {
	ID      string
	Sender  string
	Content string
	Created time.Time
}

func HandleMessage(sender string, conn *websocket.Conn) {
	localAddr := conn.LocalAddr().String()
	//read loop
	done := make(chan struct{}, 1)
	go func() {
		defer close(done)
		for {
			_, b, err := conn.ReadMessage()
			if err != nil {
				log.Println("error:", err)
				return
			}

			var msg Message
			if err := json.Unmarshal(b, &msg); err != nil {
				log.Printf("error: %v", err)
				break
			}
			from := msg.Sender
			if msg.Sender == localAddr {
				from = "you"
			}
			fmt.Printf("from: %v, message: %v \n", from, msg.Content)
		}
	}()

	//write loop
	out := make(chan Message)
	go func() {
		for {
			select {
			case <-done:
				log.Println("connection closed")
				return
			case msg, ok := <-out:
				if !ok {
					return
				}
				b, err := json.Marshal(msg)
				if err != nil {
					log.Printf("error: %v", err)
					continue
				}
				if err := conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
					log.Println("error:", err)
					return
				}
			}
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)

	for {
		fmt.Printf(">> ")
		if !scanner.Scan() {
			break
		}
		text := scanner.Text()
		msg := Message{
			Sender:  sender,
			Content: text,
			Created: time.Now(),
		}
		out <- msg
	}

	close(out)
	<-out

}
