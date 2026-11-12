package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testovoe/pkg/cfg"
)

var ConfigPath = flag.String("c", "./config.yml", "path to config")

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	flag.Parse()

	config := cfg.MustLoad(*ConfigPath)
	fmt.Println(config)

	<-sig
}
