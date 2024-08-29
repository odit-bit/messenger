package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/odit-bit/messenger/rabbit/rlog"
)

func main() {

	uri := "amqp://guest:guest@localhost:5672"
	rc := rlog.NewRabbit(uri)
	defer rc.Close()

	pub := rc.Publisher()
	defer pub.Close()

	h := rlog.NewJSONHandler(&pub, slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	// h.WithWriter(os.Stdout)
	logger := slog.New(h)
	logger2 := logger.With("start", time.Now())

	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, os.Interrupt)

	ticker := time.NewTicker(500 * time.Millisecond)

	count := 0
	stat := runtime.MemStats{}
	args := map[string]any{}

	for {
		select {
		case <-ticker.C:
			runtime.ReadMemStats(&stat)
			args["heap_alloc"] = int64(stat.HeapAlloc / 1000)
			args["next_gc"] = int64(stat.NextGC / 1000)

			count++
			if count%2 == 0 {
				logger.Info("Test Logger event", "number", count, "prime", false, "memory", args)
				continue
			} else if count%3 == 0 {
				logger.Warn("Test Logger odd", "number", count, "prime", false, "memory", args)
			} else {
				logger.Debug("Test Logger prime", "number", count, "prime", true, "memory", args)
			}

			continue
		case <-sigC:
			ticker.Stop()
			logger2.Info("shutdown", "total", count)
		}
		break
	}
	fmt.Println("total:", count)
	close(sigC)

	/// TEST

	// logger := slog.New(rlog.NewJSONHandler(os.Stdout, &rlog.Options{
	// 	Level:     nil,
	// 	RabbitURI: "",
	// }))
	// // logger := slog.New(rlog.NewIndentHandler(os.Stdout, &rlog.Options{}))

	// args := map[string]any{
	// 	"key": "value",
	// 	"num": 102020,
	// }

	// logger.Info("Test Logger", "number", 999, "Bool", true, "error", fmt.Errorf("error messages"), "Attributes", args)

	// doneC := make(chan struct{})
	// go func() {
	// 	defer close(doneC)
	// 	var wg sync.WaitGroup
	// 	for i := 0; i < 100; i++ {
	// 		wg.Add(1)
	// 		go func() {
	// 			defer wg.Done()
	// 			logger.Info("Test Logger", "number", i, "attributes", args)
	// 		}()
	// 	}
	// 	wg.Wait()
	// }()

	// <-doneC

	// logger2 := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	// logger2.Info("Test json", "key", args)

}
