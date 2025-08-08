package main

import (
	"fmt"

	"github.com/VACdotCS/kaban-go-auth-service/internal/config"
)

func main() {
	// TODO: init config
	cfg := config.MustLoad()
	fmt.Println(cfg)
	// TODO: logger init

	// TODO: init app

	// TODO: run gRPC-server of the app
}
