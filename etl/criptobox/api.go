package criptobox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/gagliardetto/solana-go/etl/http"
)

var rpc string

func SetRpc(criptoboxRpc string) {
	rpc = criptoboxRpc
}

type TxMsg struct {
	Slot      uint64 `json:"slot"`
	Signature string `json:"tx_hash"`
	FromOwner string `json:"from"`
	ToOwner   string `json:"to"`
}

type ApiResp struct {
	ErrorCode        *int64   `json:"error_code"`
	ErrorDescription *string  `json:"error_description"`
	ErrorID          *string  `json:"error_id"`
	ErrorMessage     *string  `json:"error_message"`
	Success          *bool    `json:"success"`
	Result           []string `json:"result,omitempty"`
}

type AddressList struct {
	Addresses []string `json:"addresses"`
}

func QueryAddrs(addrs []string) ([]string, error) {
	rpc, err := url.JoinPath(rpc, "/api/v1/tron/addresses")
	if err != nil {
		return nil, err
	}

	addrList := AddressList{
		Addresses: addrs,
	}
	payloadBytes, err := json.Marshal(addrList)
	if err != nil {
		return nil, err
	}

	pool := http.GetHttpClientPool()
	req, err := pool.NewRequest("GET", rpc, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := pool.Do(req)
	if err != nil {
		return nil, err
	}

	var result ApiResp
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	if !*result.Success {
		return nil, fmt.Errorf("failed to query addresses: %s", *result.ErrorMessage)
	}
	return result.Result, nil
	// return nil, fmt.Errorf("not implemented")
}

func PostTxMsgRequest(msg *TxMsg) error {
	rpc, err := url.JoinPath(rpc, "/api/v1/tron/message")
	if err != nil {
		return err
	}

	msgReq := make(map[string]string, 0)
	msgReq["tx_hash"] = msg.Signature
	payloadBytes, err := json.Marshal(msgReq)
	if err != nil {
		return err
	}
	pool := http.GetHttpClientPool()
	req, err := pool.NewRequest("POST", rpc, bytes.NewReader(payloadBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := pool.Do(req)
	if err != nil {
		return err
	}

	var result ApiResp
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	if !*result.Success {
		return fmt.Errorf("failed to post Tx msg request: %s", *result.ErrorMessage)
	}
	return nil
	// return fmt.Errorf("not implemented")
}
