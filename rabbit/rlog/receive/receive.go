package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/odit-bit/messenger/rabbit/rlog"
)

func main() {
	var level string
	flag.StringVar(&level, "level", rlog.InfoLevel, "log level severity, default 'INFO' ")
	flag.Parse()

	uri := "amqp://guest:guest@localhost:5672"
	r := rlog.NewRabbit(uri)
	defer r.Close()

	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, os.Interrupt)

	level = strings.ToLower(level)
	switch level {
	case "info":
		level = rlog.InfoLevel
	case "warn":
		level = rlog.WarnLevel
	case "debug":
		level = rlog.DebugLevel

	}

	msgC := r.Consume(level)
	for {
		select {
		case <-sigC:
		case msg, ok := <-msgC:
			if ok {
				fmt.Printf("-- %s \n", string(msg))
				continue
			}
		}
		break
	}

}
