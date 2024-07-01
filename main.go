package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/ArtemNovok/simpleRedisCl/interanl/server"
)

func main() {
	logger := setUpLogger()
	cfg := server.Config{
		Log: logger,
	}
	s := server.NewServer(cfg)
	log.Fatal(s.Start())
}

func setUpLogger() *slog.Logger {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return log
}
