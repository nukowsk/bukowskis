package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	bt "github.com/nukowsk/bukowskis/internal/types"
	"github.com/ethereum/go-ethereum/core/types"
)

// TODO: BundleSend
func HTTPSend(url string, tx *types.Transaction) (string, error) {
	request, err := bt.NewSendRawRequest(tx)
	if err != nil {
		panic(err)
	}
	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(request)

	req, err := http.NewRequest("POST", url, payloadBuf)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	jsonResp := bt.JsResponse{}
	err = json.NewDecoder(resp.Body).Decode(&jsonResp)
	if err != nil {
		return "", fmt.Errorf("Failed to decode response body: %s", err)
	}

	if jsonResp.Error != nil {
		return "", fmt.Errorf("failed %s", jsonResp.Error.Message)
	}

	if jsonResp.Result == nil {
		return "", nil
	}

	res, ok := jsonResp.Result.(string)
	if !ok {
		return "", nil
	}
	return res, nil
}

type Sender interface {
	Send(tx *types.Transaction) (string, error)
}

type HTTPSender struct {
	url string
}

func NewHTTPSender(url string) *HTTPSender {
	return &HTTPSender{
		url,
	}
}

func (h HTTPSender) Send(tx *types.Transaction) (string, error) {
	return HTTPSend(h.url, tx)
}

type MockSender struct{}

func (m MockSender) Send(tx *types.Transaction) (string, error) {
	return tx.Hash().Hex(), nil
}
