package simulation

import (
	"errors"
	"io/ioutil"
	"log"
	"math/big"
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Generator struct {
	chainID *big.Int
	keys    map[string]*keystore.Key
}

func NewGenerator(chainID *big.Int) (*Generator, error) {
	return &Generator{
		chainID,
		map[string]*keystore.Key{},
	}, nil
}

// XXX: Should probably be private
func (g *Generator) LoadWallet(path string) error {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	unlocked, err := keystore.DecryptKey([]byte(contents), "password")
	if err != nil {
		return err
	}

	g.keys[unlocked.Address.Hex()] = unlocked

	return nil
}

func (g *Generator) NewTransaction(from string, to string, args map[string]interface{}) (*types.Transaction, error) {
	fromKey, found := g.keys[from]
	if !found {
		return nil, errors.New("Don't from wallet not loaded")
	}

	amount, ok := args["amount"].(*big.Int)
	if !ok {
		amount = big.NewInt(100)
	}

	gasLimit, ok := args["gasLimit"].(uint64)
	if !ok {
		gasLimit = uint64(21000)
	}

	gasPrice, ok := args["gasPrice"].(*big.Int)
	if !ok {
		gasPrice = big.NewInt(600)
	}

	nonce := uint64(0)
	data := []byte{}

	tx := types.NewTransaction(
		nonce,
		common.HexToAddress(to),
		amount,
		gasLimit,
		gasPrice,
		data)

	signer := types.NewEIP155Signer(g.chainID)
	signedTx, err := types.SignTx(tx, signer, fromKey.PrivateKey)

	if err != nil {
		return nil, errors.New("failed to sign transaction")
	}

	return signedTx, nil
}

func (g *Generator) NumAccounts() int {
	return len(g.keys)
}

// XXX: Use either keys or accounts naming
func (g *Generator) RandAccount() string {
	rand.Seed(time.Now().Unix())
	choice := rand.Intn(len(g.keys))
	i := 0
	for key, _ := range g.keys {
		if i == choice {
			return key
		}
		i++
	}
	return "" // XXX: Probably return an error
}

func (g *Generator) RandomPair() (string, string) {
	if g.NumAccounts() < 2 {
		log.Fatalf("Require at least two accounts to run simulation")
	}
	from := g.RandAccount()
	var to string
	for i := 0; i < g.NumAccounts(); i++ {
		to = g.RandAccount()
		if from != to {
			break
		}
	}

	return from, to
}

func (g *Generator) Next() *types.Transaction {
	from, to := g.RandomPair()
	args := map[string]interface{}{
		"amount": int64(1000),
	}
	tx, err := g.NewTransaction(from, to, args)
	if err != nil {
		panic(err)
	}

	return tx
}
