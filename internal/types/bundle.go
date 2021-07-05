package types

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

/*
Create a minimal set of helpers functions for producing flashbots bundles
*/

var (
	signerKey      = "privateKey"
	bundleEndpoint = "https://relay-goerli.flashbots.net"
	rpcEndpoint    = ""
	method         = "eth_sendBundle"
)

type JsonRpc struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int64         `json:"id"`
}

type Bundle struct {
	Header string
	Body   []byte
}

func signHash(hashString string) []byte {
	// XXX: Is this correct?
	// docs suggest sig, err := crypto.Sign(crypto.Keccak256([]byte("\x19Ethereum Signed Message:\n"+strconv.Itoa(len(hashedBody))+hashedBody)), pk)
	return crypto.Keccak256([]byte("\x19Ethereum Signed Message:\n66" + hashString))
}

func BundleAndSend(tx *types.Transaction) {
	client, err := ethclient.Dial(rpcEndpoint)
	if err != nil {
		log.Fatalf("Failed to connect to rpcEncpoint %s\n", err)
	}

	// get the current block number
	blockNumber, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Fatalf("Couldn't get block number: %s\n", err)
	}

	bundle, err := NewBundle(signerKey, tx, blockNumber)
	if err != nil {
		log.Fatalf("No bundle: %s\n", err)
	}

	bundleClient := &http.Client{}
	req, err := http.NewRequest("POST", bundleEndpoint, bytes.NewBuffer(bundle.Body))
	if err != nil {
		log.Fatalf("Failed to construct request: %s\n", err)
	}

	res, err := bundleClient.Do(req)
	if err != nil {
		log.Fatalf("Failed to submit bundle: %s\n", err)
	}

	log.Printf("Success: %+v\n", res)
}

func NewBundle(signer string, tx *types.Transaction, blockNumber uint64) (*Bundle, error) {
	// Turn into a jsonRPCREquest
	hash, err := HexEncodeTransaction(tx)
	if err != nil {
		return nil, err
	}
	rpcReq := JsonRpc{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  []interface{}{hash, fmt.Sprintf("%x", blockNumber)},
		ID:      1,
	}

	marshalled, err := json.Marshal(rpcReq)
	if err != nil {
		return nil, fmt.Errorf("Failed to marhsall json %s\n", err)
	}

	// sign the payload
	ecdsaPrivateKey, err := crypto.HexToECDSA(signer)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode private key: %s\n", err)
	}

	publicKey := ecdsaPrivateKey.PublicKey
	// XXX: convert this to a string
	signerAddr := hexutil.Encode(crypto.PubkeyToAddress(publicKey).Bytes())

	signatureBytes, err := crypto.Sign(
		signHash(hexutil.Encode(crypto.Keccak256(marshalled))),
		ecdsaPrivateKey)

	if err != nil {
		return nil, err
	}
	signature := hexutil.Encode(signatureBytes)

	header := signerAddr + ":" + signature

	return &Bundle{
		Body:   marshalled,
		Header: header,
	}, nil
}
