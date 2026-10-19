package main

import (
	"os"

	"github.com/defos-ai/commander-web-agent/cli"
)

func main() { os.Exit(cli.Main(os.Args[1:], os.Stderr)) }
