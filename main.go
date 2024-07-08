package main

import (
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/ArtemNovok/simpleRedisCl/internal/server"
)

const (
	lvlProd  = "PROD"
	lvlDebug = "DEV"
)

func main() {
	addr := flag.String("listenAddr", server.DefaultAddress, "listen address of the server")
	lvl := flag.String("loglvl", lvlDebug, "level of the logs ('PROD', 'DEV')")
	password := flag.String("password", "", "password that used to connect to a server")
	flag.Parse()
	logger := setUpLogger(*lvl)
	cfg := server.Config{
		Log:        logger,
		ListenAddr: *addr,
		Password:   *password,
	}
	s := server.NewServer(cfg)
	log.Fatal(s.Start())
}

func setUpLogger(lvl string) *slog.Logger {
	switch lvl {
	case lvlProd:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case lvlDebug:
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	default:
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
}
