package many

import (
	"fmt"
	"log/slog"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

type TxInfo struct {
	Arguments struct {
		From   string   `json:"from"`
		To     string   `json:"to"`
		Amount int64    `json:"amount"` // Talib is not following the specification. This should be a BigInt
		Symbol string   `json:"symbol"`
		Memo   []string `json:"memo"`
	} `json:"argument"`
}

func GetTxInfo(r *resty.Client, hash string) (*TxInfo, error) {
	req := r.R().SetPathParam("thash", hash).SetResult(&TxInfo{})
	resp, err := req.Get("neighborhoods/{neighborhood}/transactions/{thash}")
	if err != nil {
		slog.Error("Error getting MANY tx info", "error", err)
		return nil, err
	}

	txInfo := resp.Result().(*TxInfo)
	if txInfo == nil || (txInfo != nil && txInfo.Arguments.From == "") {
		slog.Error("Error unmarshalling MANY tx info")
		return nil, fmt.Errorf("error unmarshalling MANY tx info")
	}
	slog.Debug("MANY tx info", "txInfo", txInfo)
	return txInfo, nil
}

func CheckTxInfo(txInfo *TxInfo, itemUUID uuid.UUID, manifestAddr string) error {
	// Check the MANY transaction `To` address
	if txInfo.Arguments.To != IllegalAddr {
		return fmt.Errorf("invalid MANY tx `to` address: %s", txInfo.Arguments.To)
	}

	// Check the MANY transaction `Memo`
	if len(txInfo.Arguments.Memo) != 2 {
		return fmt.Errorf("invalid MANY Memo length: %d", len(txInfo.Arguments.Memo))
	}

	// Check the MANY transaction UUID
	txUUID, err := uuid.Parse(txInfo.Arguments.Memo[0])
	if err != nil {
		return fmt.Errorf("invalid MANY tx UUID: %s", txInfo.Arguments.Memo[0])
	}

	// Check the Manifest destination address
	if txInfo.Arguments.Memo[1] != manifestAddr {
		return fmt.Errorf("invalid manifest destination address: %s", txInfo.Arguments.Memo[1])
	}

	// Check the MANY transaction UUID matches the work item UUID
	if txUUID != itemUUID {
		return fmt.Errorf("MANY tx UUID does not match work item UUID: %s, %s", txUUID, itemUUID)
	}

	return nil
}
