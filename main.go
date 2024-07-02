package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/ArtemNovok/simpleRedisCl/interanl/client"
	"github.com/ArtemNovok/simpleRedisCl/interanl/server"
)

func main() {
	logger := setUpLogger()
	cfg := server.Config{
		Log: logger,
	}
	s := server.NewServer(cfg)
	go func() {
		log.Fatal(s.Start())
	}()
	time.Sleep(1 * time.Second)
	cl := client.New("localhost:6666")
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("key%v", i)
		val := fmt.Sprintf("val%v", i)
		go func() {
			err := cl.Set(context.Background(), key, val)
			if err != nil {
				fmt.Println(err)
			}
		}()
	}
	select {}
}

func setUpLogger() *slog.Logger {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return log
}
