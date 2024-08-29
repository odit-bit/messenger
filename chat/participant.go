package chat

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// PARTICIPANT

type Participant struct {
	conn     *websocket.Conn
	messages chan []byte
	chat     *Chat
}

// read message from conn and send into chat rooms
func (p *Participant) ReadLoop() {
	defer func() {
		p.chat.leaveC <- p
		log.Printf("participant leave:%v \n", p.conn.RemoteAddr())
		p.conn.Close()
	}()

	p.conn.SetReadDeadline(time.Now().Add(pongWait))
	// set pONG
	p.conn.SetPongHandler(func(appData string) error {
		p.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, b, err := p.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		p.chat.brodcastC <- b
	}
}

// write message from chat room into conn
func (p *Participant) WriteLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		p.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-p.messages:
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				p.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			p.conn.WriteMessage(websocket.BinaryMessage, msg)

		case <-ticker.C:
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := p.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
