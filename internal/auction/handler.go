package auction

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/nukowsk/bukowskis/internal/sender"
	st "github.com/nukowsk/bukowskis/internal/store"
	bt "github.com/nukowsk/bukowskis/internal/types"
	"github.com/ethereum/go-ethereum/core/types"
)

type Handler struct {
	proxy     http.Handler
	processTx func(*types.Transaction) (string, error)
}

func NewHandler(
	gasGetter GasGetter,
	store st.Store,
	sender sender.Sender,
	proxy http.Handler) *Handler {
	processTx := genProcessTx(gasGetter, store, sender)
	return &Handler{
		proxy:     proxy,
		processTx: processTx,
	}
}

func (h *Handler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	jsr, err := bt.ParseRequest(req)
	if err != nil {
		log.Printf("Error: parsing request body: %v\n", err)
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Request: %+v\n", jsr.Method)
	if jsr.Method == "eth_sendRawTransaction" ||
		jsr.Method == "eth_sendTransaction" ||
		jsr.Method == "eth_sendRawTransaction_reserve" ||
		jsr.Method == "eth_sendTransaction_reserve" {

		tx, err := bt.ExtractTransaction(jsr)
		if err != nil {
			log.Printf("Error: extracting transaction %s\n", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("Received: %s\n", tx.Hash().Hex())
		var response bt.JsResponse
		result, err := h.processTx(tx)
		if err != nil {
			log.Printf("Failed: %s\n%s\n", tx.Hash().Hex(), err)
			response = bt.NewJsError(-1, err.Error())
		} else {
			log.Printf("Success: %s\n", tx.Hash().Hex())
			response = bt.JsResponse{
				Result: result,
			}
		}

		err = json.NewEncoder(res).Encode(response)
		if err != nil {
			log.Fatal("Failed to encode json response")
		}
	} else {
		log.Printf("Proxy to vanilla: %+v\n", jsr.Method)
		h.proxy.ServeHTTP(res, req)
	}
}

func genProcessTx(
	gasGetter GasGetter,
	store st.Store,
	sender sender.Sender) func(*types.Transaction) (string, error) {
	return func(tx *types.Transaction) (string, error) {
		minGas := gasGetter.FastPrice()
		if tx.GasPrice().Cmp(minGas) == -1 {
			return "", fmt.Errorf("Gas too low")
		}

		entry, err := st.NewLogEntry(tx)
		if err != nil {
			return "", fmt.Errorf("Error: creating log entry: %s", err)
		}

		err = store.Save(&entry)
		if err != nil {
			return "", fmt.Errorf("Error: failed to store transaction %s", err)
		}

		result, err := sender.Send(tx)
		if err != nil {
			return "", fmt.Errorf("Error: failed to submit transaction %s", err)
		}

		return result, err
	}
}
