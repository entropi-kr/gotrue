package main

import (
	"log"

	"gitlab.com/entropi-tech/gotrue/cmd"
)

func main() {
	if err := cmd.RootCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}
