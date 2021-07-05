package main

import (
	"context"
	"log"
	"net/url"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/nukowsk/bukowskis/internal/auction"
	"github.com/nukowsk/bukowskis/internal/sender"
	"github.com/nukowsk/bukowskis/internal/simulation"
	"github.com/nukowsk/bukowskis/internal/store"
)

func main() {
	log.Print("starting action server")

	projectId := os.Getenv("BUKOWSKIS_PROJECT_ID")
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectId)
	if err != nil {
		log.Fatalf("Fatal firebase error :%s", err)
	}
	defer client.Close()

	store, err := store.NewFirestore(projectId)
	// store, err := store.NewLocal()
	if err != nil {
		log.Fatalf("Couldn't initialize store: %v\n", store)
	}

	bidderURL, err := url.Parse(os.Getenv("BUKOWSKIS_BIDDER_URL"))
	if err != nil {
		log.Fatalf("Set environment variable BUKOWSKIS_BIDDER_URL")
	}

	log.Printf("Starting bidder proxy: %s", bidderURL)

	vanillaURL, err := url.Parse(os.Getenv("BUKOWSKIS_VANILLA_URL"))
	if err != nil {
		log.Fatalf("Set environment variable BUKOWSKIS_VANILLA_URL")
	}

	log.Printf("Starting vanilla proxy: %s", vanillaURL)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	url := "https://ethgasstation.info/json/ethgasAPI.json"
	gasService, err := auction.NewGasService(url)
	if err != nil {
		log.Fatalf("failed to initialize gas service: %s\n", err)
	}
	simulate := os.Getenv("BUKOWSKIS_SIMULATE") == "true"
	if simulate {
		auctionAddr := os.Getenv("BUKOWSKIS_AUCTION_ADDR")
		if auctionAddr == "" {
			log.Fatalln("Required env BUKOWSKIS_AUCTION_ADDR")
		}

		keysDir := os.Getenv("BUKOWSKIS_KEYS_DIR")
		if keysDir == "" {
			log.Fatalln("Required env BUKOWSKIS_KEYS_DIR")
		}
		log.Printf("Loading keys from: %s\n", keysDir)

		service, err := simulation.NewService(
			auctionAddr,
			vanillaURL.String(),
			keysDir,
			1*time.Minute)
		if err != nil {
			log.Fatalf("Failed to initiate simulation: %s\n", err)
		}

		go service.Run(-1)
	}

	proxy := auction.NewProxy(vanillaURL)
	sender := sender.NewHTTPSender(bidderURL.String())
	server, err := auction.NewAuctionService(
		port,
		proxy,
		sender,
		store,
		gasService)
	if err != nil {
		log.Fatalf("Failed to initialize auction server: %s\n", err)
	}

	log.Printf("listening on port %s", port)
	go gasService.Run()
	server.Run()
}
