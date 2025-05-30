//go:build !api

package main

import (
	"github.com/JackBekket/reflexia/cmd/cli"
	"github.com/JackBekket/reflexia/pkg/config"
)

func main() {
	cfg := config.NewConfig()

	cli.Run(cfg)
}
