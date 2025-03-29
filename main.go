package main

import (
	"RenGO/engine"
	"log"
)

func main() {
	game := engine.NewEngine(1280, 720, 5)

	if err := game.Run(); err != nil {
		log.Fatal(err)
	}

}
