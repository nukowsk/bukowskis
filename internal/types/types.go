package types

// XXX: rename this pacakge to utils
import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

func ParseTransaction(hexStr string) (*types.Transaction, error) {
	data := common.FromHex(hexStr)
	var tx types.Transaction
	err := tx.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func UnsignTransaction(tx *types.Transaction) *types.Transaction {
	newTx := types.NewTransaction(tx.Nonce(), *(tx.To()), tx.Value(), tx.Gas(), tx.GasPrice(), tx.Data())

	return newTx
}

func HexEncodeTransaction(tx *types.Transaction) (string, error) {
	data, err := tx.MarshalBinary()
	if err != nil {
		return "", err
	}
	return hexutil.Encode(data), nil
}

func PPTransaction(tx *types.Transaction) string {
	fmtStr := `
id: %s
to: %s
amount: %d
gas: %d
	`
	return fmt.Sprintf(
		fmtStr,
		tx.Hash().Hex(),
		hexutil.Encode(tx.To().Bytes()),
		tx.Value(),
		tx.Gas(),
	)
}
