package simulation

import (
	"context"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	"github.com/nukowsk/bukowskis/internal/sender"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Service struct {
	generator   *Generator
	fin         chan struct{}
	freq        time.Duration
	auctionAddr string
}

func NewService(
	auctionAddr string,
	vanillaAddr string,
	keysDir string,
	freq time.Duration,
) (*Service, error) {
	client, err := ethclient.Dial(vanillaAddr)
	if err != nil {
		return nil, err
	}

	// get chain Id
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	generator, err := NewGenerator(chainID)

	if err != nil {
		return nil, err
	}

	log.Printf("keysDir: %s\n", keysDir)
	files, err := ioutil.ReadDir(keysDir)
	if err != nil {
		log.Fatalln("Couldn't read keys dir")
	}

	for _, f := range files {
		path := filepath.Join(keysDir, f.Name())
		err = generator.LoadWallet(path)
		if err != nil {
			log.Fatalf("Failed to load wallet: %s\n", err)
		}
	}

	return &Service{
		generator:   generator,
		fin:         make(chan struct{}),
		auctionAddr: auctionAddr,
		freq:        freq,
	}, nil
}

func (s *Service) Run(maxIter int) {
	timer := time.NewTicker(s.freq)
	count := 0
	for {
		select {
		case <-timer.C:
			if maxIter != -1 && count == maxIter {
				break
			}
			count += 1
			s.simulate()
		case _ = <-s.fin:
			timer.Stop()
			break
		}
	}
}
func (s *Service) simulate() {
	tx := s.generator.Next()
	log.Printf("Simulator generated: %s\n", tx.Hash().Hex())
	_, err := sender.HTTPSend(s.auctionAddr, tx)
	if err != nil {
		log.Printf("Transaction was not accepted: %s\n", err)
	}
}

func (s *Service) Stop() {
	close(s.fin)
}
