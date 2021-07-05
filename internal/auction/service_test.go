package auction

import (
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nukowsk/bukowskis/internal/sender"
	"github.com/nukowsk/bukowskis/internal/simulation"
	"github.com/nukowsk/bukowskis/internal/store"
)

func TestInitialization(t *testing.T) {
	var (
		port       = "8080"
		proxy      = MockProxy{}
		thissender = sender.MockSender{}
		store, _   = store.NewLocal()
	)
	server, err := NewAuctionService(
		port,
		proxy,
		thissender,
		store,
		&MockGasGetter{
			price: big.NewInt(400),
		})

	if err != nil {
		t.Errorf("Auction service init failed failed %s\n", err)
	}

	go func() {
		server.Run()
	}()

	// Wait for server to start
	time.Sleep(1 * time.Second)

	chainID := big.NewInt(999)
	generator, err := simulation.NewGenerator(chainID)
	if err != nil {
		panic(err)
	}

	// Load wallets
	keysDir := os.Getenv("BUKOWSKIS_KEYS_DIR")
	if keysDir == "" {
		log.Fatalln("Required env BUKOWSKIS_KEYS_DIR")
	}

	keysDir = filepath.Join("..", "..", keysDir)
	files, err := ioutil.ReadDir(keysDir)
	if err != nil {
		log.Fatalf("Loading wallets: %s\n", err)
	}

	for _, f := range files {
		path := filepath.Join(keysDir, f.Name())
		err = generator.LoadWallet(path)
		if err != nil {
			log.Fatalf("Failed to load wallet: %s\n", err)
		}
	}

	to, from := generator.RandomPair()
	args := map[string]interface{}{
		"amount": int64(1000),
	}

	tx, err := generator.NewTransaction(to, from, args)
	if err != nil {
		log.Fatalf("Failed to generate transaction: %s\n", err)
	}

	result, err := sender.HTTPSend("http://localhost:8080", tx)
	if err != nil {
		t.Fatalf("Failed to submit transaction: %s\n", err)
	}

	if result != tx.Hash().Hex() {
		errorstr := `
Transaction didn't match:
	expected: %s
	got: %s
		`
		t.Errorf(errorstr, tx.Hash().Hex(), result)
	}

	args2 := map[string]interface{}{
		"amount":   int64(1000),
		"gasPrice": big.NewInt(1),
	}
	tx, err = generator.NewTransaction(to, from, args2)
	if err != nil {
		log.Fatalf("Failed to generate transaction: %s\n", err)
	}

	result, err = sender.HTTPSend("http://localhost:8080", tx)
	if err == nil {
		t.Fatalf("Transaction should fail: %s\n", result)
	}

	server.Stop()
}
