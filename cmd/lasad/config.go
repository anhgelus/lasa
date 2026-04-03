package main

import (
	_ "embed"
	"fmt"
	"os"

	"tangled.org/anhgelus.world/lasa/cmd/internal"
)

//go:embed default.toml
var defaultConfig []byte

func handleGenConfigHelp() {
	internal.Usage(
		`lasad gen-config`,
		"Generate a new config. Can destroy yours!",
		nil,
		flags,
		[]string{
			"lasad gen-config\t-\tgenerate a new config",
		},
	)
	if !help {
		os.Exit(1)
	}
}

func handleGenConfig(args []string) {
	if len(args) != 0 || help {
		handleGenConfigHelp()
		return
	}
	fmt.Println("writing default file at", configPath)
	err := os.WriteFile(configPath, defaultConfig, 0640)
	if err != nil {
		panic(err)
	}
}
