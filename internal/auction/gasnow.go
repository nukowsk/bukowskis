package auction

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/big"
	"net/http"
	"sync"
	"time"
)

// Create an intergration with GasNow
// https://ethgasstation.info/json/ethgasAPI.json

type GasGetter interface {
	FastPrice() *big.Int
}

type gasNowResponse struct {
	Fast float64 `json:fast`
}

func pollGasPrice(url string) (*big.Int, error) {
	client := http.Client{}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res gasNowResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode response body: %s", err)
	}

	k := math.Round(res.Fast)
	i := big.NewInt(int64(k))
	return i, nil
}

type GasService struct {
	url   string
	mx    sync.Mutex // XXX: Switch to RWLock
	fin   chan interface{}
	price *big.Int
}

// XXX: Might want to seperate construction from intialization
// Return error
func NewGasService(url string) (*GasService, error) {
	price, err := pollGasPrice(url)

	if err != nil {
		return nil, err
	}
	return &GasService{
		url:   url,
		fin:   make(chan interface{}),
		price: price,
	}, nil
}

func (g *GasService) Run() {
	log.Println("running gas service")
	timer := time.NewTicker(60 * time.Second) // XXX: Configure
	for {
		select {
		case <-timer.C:
			g.updateGas()
		case _ = <-g.fin:
			timer.Stop()
			break
		}
	}
}

func (g *GasService) updateGas() {
	g.mx.Lock()
	defer g.mx.Unlock()
	price, err := pollGasPrice(g.url)
	if err != nil {
		log.Printf("Failed to update min gas %s\n", err)
	}

	log.Printf("New gas %d\n", price) // XXX: Remove

	g.price = price
}

func (g *GasService) Stop() {
	close(g.fin)
}

func (g *GasService) FastPrice() *big.Int {
	g.mx.Lock()
	defer g.mx.Unlock()
	return g.price
}

type MockGasGetter struct {
	price *big.Int
}

func (m *MockGasGetter) FastPrice() *big.Int {
	return m.price
}
