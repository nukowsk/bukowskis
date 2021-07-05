package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/core/types"
)

type JsRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params,omitempty"`
}

// See: http://www.jsonrpc.org/specification#error_object
type JsError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *JsError) Error() string {
	return strconv.Itoa(e.Code) + ":" + e.Message
}

// See: http://www.jsonrpc.org/specification#response_object
type JsResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *JsError    `json:"error,omitempty"`
	ID      int         `json:"id"`
}

func ExecutionError() JsResponse {
	return JsResponse{
		JSONRPC: "2",
		ID:      1,
		Error: &JsError{
			Code:    3,
			Message: "Execution error",
		},
	}
}

func NewSendRawRequest(tx *types.Transaction) (JsRequest, error) {
	hash, err := HexEncodeTransaction(tx)
	if err != nil {
		return JsRequest{}, err
	}
	return JsRequest{
		JSONRPC: "2.0",
		Method:  "eth_sendRawTransaction",
		Params:  []interface{}{hash},
	}, nil
}

// XXX: This will NOT mutate the http request
func ParseRequest(request *http.Request) (JsRequest, error) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return JsRequest{}, err
	}

	var jsr JsRequest
	err = json.Unmarshal([]byte(body), &jsr)
	if err != nil {
		return JsRequest{}, err
	}

	request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	return jsr, nil
}

func NewJsError(code int, message string) JsResponse {
	return JsResponse{
		JSONRPC: "2.0",
		Error: &JsError{
			Code:    code,
			Message: message,
		},
		ID: 1,
	}
}

// pre-condition; this is a eth_sendRawTransaction
func ExtractTransaction(req JsRequest) (*types.Transaction, error) {
	if len(req.Params) != 1 {
		return nil, fmt.Errorf("Invalid Request, too many params")
	}
	str, ok := req.Params[0].(string)
	if !ok {
		return nil, fmt.Errorf("Invalid Request, should be a string")
	}

	tx, err := ParseTransaction(str)
	if err != nil {
		return nil, fmt.Errorf("Invalid Request, not a couldn't decode transaction")
	}

	return tx, nil
}
