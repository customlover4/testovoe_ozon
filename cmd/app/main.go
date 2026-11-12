package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"testovoe/pkg/cfg"
	"testovoe/pkg/logger"
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

	<-sig
}
