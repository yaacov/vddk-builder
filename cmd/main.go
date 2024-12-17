package main

import (
	"vddk-builder/pkg/config"
	"vddk-builder/pkg/server"
)

func main() {
	cfg := config.LoadConfig()
	server.StartServer(cfg)
}
