package main

import (
	"context"
	"errors"
	"flag"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testovoe/internal/web"
	"testovoe/pkg/cfg"
	"testovoe/pkg/logger"
	"time"
)

var ConfigPath = flag.String("c", "./config/config.yml", "path to config")

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	flag.Parse()

	config := cfg.MustLoad(*ConfigPath)

	log := logger.New(
		config.Logger.Level, config.Debug,
	)

	log.Info().Msg("Config proceed is done.")
	log.Info().Msg("New logger created")

	addr := config.Web.Addr + ":" + config.Web.Port

	server := http.Server{
		Handler: web.NewRouter(config.Debug),

		Addr:              addr,
		WriteTimeout:      config.Web.WriteTimeout,
		ReadTimeout:       config.Web.ReadTimeout,
		ReadHeaderTimeout: config.Web.ReadHeaderTimeout,
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal().Err(err).Msg("Can't create listener")
	}

	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("Server shutdown with err")
		}
	}()

	log.Info().Str("addr", addr).Msg("Server started")

	<-sig

	log.Info().Msg("Start shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Server shutdown with err")
	}
}
