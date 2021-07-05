package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/nukowsk/bukowskis/internal/types"
)

func validRequest(req types.JsRequest) bool {
	return req.Method == "eth_sendRawTransaction"
}

func main() {
	log.Println("Starting Monopolistic Bidder")

	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		jsr, err := types.ParseRequest(req)
		if err != nil {
			log.Printf("Error parsing request body: %v\n", err)
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("Request: %#v\n", jsr)
		if !validRequest(jsr) {
			log.Printf("Invalid request: %s\n", jsr)
		}

		tx, err := types.ExtractTransaction(jsr)
		var resp types.JsResponse
		if err != nil {
			// what's the correct error code here?
			resp = types.NewJsError(0, "Invalid payload")
		} else {
			resp = types.JsResponse{
				Result: tx.Hash().Hex(),
			}
		}

		err = json.NewEncoder(res).Encode(resp)
		if err != nil {
			log.Printf("Encoding error %s", err)
		}
	})

	// parse the address here
	url, err := url.Parse(os.Getenv("BUKOWSKIS_BIDDER_URL"))
	if err != nil {
		log.Fatal("BUKOWSKIS_BIDDER_URL required")
	}

	port := url.Port()

	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
