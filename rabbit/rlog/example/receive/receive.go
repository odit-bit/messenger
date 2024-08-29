package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/odit-bit/messenger/rabbit/rlog"
)

func main() {
	// var level string
	// flag.StringVar(&level, "level", "", "log level severity, default 'INFO'")
	// flag.Parse()

	// if level == "" {
	// 	log.Fatal("need flag, flag come first before args 'cmd -level=value arg'")
	// }
	if len(os.Args) < 1 {
		log.Fatal("need more argument")
	}
	level := os.Args[1]

	queue := strings.ToLower(level)
	switch queue {
	case "info":
		queue = rlog.InfoLevel
	case "warn":
		queue = rlog.WarnLevel
	case "debug":
		queue = rlog.DebugLevel
	default:
		log.Fatalf("unknown flag %s \n", level)
	}

	uri := "amqp://guest:guest@localhost:5672"
	r := rlog.NewRabbit(uri)
	defer r.Close()

	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, os.Interrupt)

	msgC := r.Consumer(queue)
	for {
		select {
		case <-sigC:
		case msg, ok := <-msgC:
			if ok {
				fmt.Printf("-- %s \n", string(msg.Body))
				msg.Ack(false)
				continue
			}
		}
		break
	}

}
