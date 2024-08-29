package chat

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/segmentio/ksuid"
)

var (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

// CHAT

type Chat struct {
	Name        string
	participant map[*Participant]bool
	brodcastC   chan []byte
	leaveC      chan *Participant

	messageStore map[string]Message
}

func NewChat(name string) *Chat {
	return &Chat{
		Name:         name,
		participant:  map[*Participant]bool{},
		brodcastC:    make(chan []byte),
		leaveC:       make(chan *Participant),
		messageStore: map[string]Message{},
	}
}

// func (c *Chat) Add(conn *websocket.Conn) {
// 	p := Participant{
// 		conn:     conn,
// 		messages: make(chan []byte),
// 		chat:     c,
// 	}
// 	go p.ReadLoop()
// 	go p.WriteLoop()
// 	c.participant[&p] = true
// }

func (c *Chat) AddParticipant(p *Participant) {
	go p.ReadLoop()
	go p.WriteLoop()
	c.participant[p] = true
}

func (c *Chat) Run() {
	var wg sync.WaitGroup
	for {
		select {
		case msg := <-c.brodcastC:
			wg.Add(1)
			go func() {
				defer wg.Done()
				var message Message
				if err := json.Unmarshal(msg, &message); err != nil {
					log.Println(err)
				}
				message.ID = ksuid.New().String()
				c.messageStore[message.ID] = message
			}()
			for p := range c.participant {
				select {
				case p.messages <- msg:
				default:
					close(p.messages)
					delete(c.participant, p)
				}
			}
			wg.Wait()
		case p := <-c.leaveC:
			if _, ok := c.participant[p]; ok {
				delete(c.participant, p)
				close(p.messages)
			}

		}
	}

}

func (c *Chat) Messages() []Message {
	messages := []Message{}
	for _, v := range c.messageStore {
		messages = append(messages, v)
	}

	return messages
}
