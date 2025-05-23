//go:build api

package main

import (
	"github.com/JackBekket/reflexia/cmd/api"
	"github.com/JackBekket/reflexia/pkg/config"
)

func main() {
	cfg := config.NewConfig()

	api.Run(cfg)
}
