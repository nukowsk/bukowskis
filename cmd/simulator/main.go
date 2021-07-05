package main

import (
	"log"
	"os"
	"time"

	"github.com/nukowsk/bukowskis/internal/simulation"
)

func main() {
	vanillaAddr := os.Getenv("BUKOWSKIS_VANILLA_URL")
	if vanillaAddr == "" {
		log.Fatalln("Required env BUKOWSKIS_VANILLA_URL")
	}

	auctionAddr := os.Getenv("BUKOWSKIS_AUCTION_ADDR")
	if auctionAddr == "" {
		log.Fatalln("Required env BUKOWSKIS_AUCTION_ADDR")
	}

	log.Printf("Sending to: %s\n", auctionAddr)

	keysDir := os.Getenv("BUKOWSKIS_KEYS_DIR")
	if keysDir == "" {
		log.Fatalln("Required env BUKOWSKIS_KEYS_DIR")
	}
	log.Printf("Loading keys from: %s\n", keysDir)

	service, err := simulation.NewService(
		auctionAddr,
		vanillaAddr,
		keysDir,
		1*time.Second)
	if err != nil {
		log.Fatalf("Failed to initiate simulation: %s\n", err)
	}

	service.Run(10)
	log.Println("Simulation complete")
}
