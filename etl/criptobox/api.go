package criptobox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

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

func RegisterTxMsg(msg *TxMsg) error {
	addrs, err := QueryAddrs([]string{msg.FromOwner, msg.ToOwner})
	if err != nil {
		return err
	}
	if len(addrs) <= 0 {
		return fmt.Errorf("non-related Tx")
	}

	for {
		err = PostTxMsgRequest(msg)
		if err != nil {
			log.Printf("ERROR %d %s %v", msg.Slot, msg.Signature, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}

	log.Printf("WARN %d from: %s to: %s %s", msg.Slot,
		formatAddress(msg.FromOwner), formatAddress(msg.ToOwner), msg.Signature)
	return nil
}

func QueryAddrs(addrs []string) ([]string, error) {
	rpc, err := url.JoinPath(rpc, "/api/v1/solana/addresses")
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
		log.Printf("ERROR %v %s", string(resp), err)
		return nil, err
	}
	if !*result.Success {
		return nil, fmt.Errorf("failed to query addresses: %s", *result.ErrorMessage)
	}
	return result.Result, nil
	// return nil, fmt.Errorf("not implemented")
}

func PostTxMsgRequest(msg *TxMsg) error {
	rpc, err := url.JoinPath(rpc, "/api/v1/solana/message")
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
		if strings.Contains(*result.ErrorMessage, "tx_hash already exists") {
			return nil
		}
		return fmt.Errorf("failed to post Tx msg request: %s", *result.ErrorMessage)
	}
	return nil
	// return fmt.Errorf("not implemented")
}

// formatAddress shortens an address for display
func formatAddress(address string) string {
	// if len(address) <= 12 {
	// 	return address
	// }
	// return address[:4] + "..." + address[len(address)-4:]
	if len(address) <= 6 {
		return address
	}
	return address[len(address)-6:]
}
