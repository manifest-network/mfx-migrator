package many

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Arguments struct {
	From   string   `json:"from"`
	To     string   `json:"to"`
	Amount string   `json:"amount"`
	Symbol string   `json:"symbol"`
	Memo   []string `json:"memo"`
}

type MultisigSubmitTransaction struct {
	Arguments Arguments `json:"argument"`
}

type MultisigSubmitTransactionArguments struct {
	Transaction MultisigSubmitTransaction `json:"transaction"`
}

type TxInfo struct {
	Method    string          `json:"method"`
	Arguments json.RawMessage `json:"argument"`
}

func GetTxInfo(r *resty.Client, hash string) (*Arguments, error) {
	req := r.R().SetPathParam("thash", hash).SetResult(&TxInfo{})
	resp, err := req.Get("neighborhoods/{neighborhood}/transactions/{thash}")
	if err != nil {
		return nil, errors.WithMessage(err, "error unmarshalling MANY tx info")
	}

	txInfo := resp.Result().(*TxInfo)
	if txInfo == nil {
		return nil, fmt.Errorf("response not a MANY tx info")
	}

	switch txInfo.Method {
	case "ledger.send":
		var args Arguments
		if err := json.Unmarshal(txInfo.Arguments, &args); err != nil {
			return nil, errors.WithMessage(err, "error unmarshalling ledger.send tx arguments")
		}
		return &args, nil
	case "account.multisigSubmitTransaction":
		var args MultisigSubmitTransactionArguments
		if err := json.Unmarshal(txInfo.Arguments, &args); err != nil {
			return nil, errors.WithMessage(err, "error unmarshalling multisigSubmitTransaction tx arguments")
		}
		return &args.Transaction.Arguments, nil
	default:
		return nil, fmt.Errorf("unsupported MANY tx method: %s", txInfo.Method)
	}
}

func CheckTxInfo(txArgs *Arguments, itemUUID uuid.UUID, manifestAddr string) error {
	// Check the MANY transaction `To` address
	if txArgs.To != IllegalAddr {
		return fmt.Errorf("invalid MANY tx `to` address: %s", txArgs.To)
	}

	// Check the MANY transaction `Memo`
	if len(txArgs.Memo) != 2 {
		return fmt.Errorf("invalid MANY Memo length: %d", len(txArgs.Memo))
	}

	// Check the MANY transaction UUID
	txUUID, err := uuid.Parse(txArgs.Memo[0])
	if err != nil {
		return errors.WithMessagef(err, "invalid MANY tx UUID: %s", txArgs.Memo[0])
	}

	// Check the Manifest destination address
	if txArgs.Memo[1] != manifestAddr {
		return fmt.Errorf("invalid manifest destination address: %s", txArgs.Memo[1])
	}

	// Check the MANY transaction UUID matches the work item UUID
	if txUUID != itemUUID {
		return fmt.Errorf("MANY tx UUID does not match work item UUID: %s, %s", txUUID, itemUUID)
	}

	return nil
}
