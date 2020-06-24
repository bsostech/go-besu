package types

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/bsostech/go-besu/privacy"
)

// PrivateReceipt represents the results of a transaction.
type PrivateReceipt struct {
	// Consensus fields: These fields are defined by the Yellow Paper
	PostState []byte `json:"root"`
	Status    uint64 `json:"status"`
	// CumulativeGasUsed uint64       `json:"cumulativeGasUsed" gencodec:"required"`
	Bloom types.Bloom  `json:"logsBloom"         gencodec:"required"`
	Logs  []*types.Log `json:"logs"              gencodec:"required"`

	// Implementation fields: These fields are added by geth when processing a transaction.
	// They are stored in the chain database.
	TxHash          common.Hash    `json:"transactionHash" gencodec:"required"`
	ContractAddress common.Address `json:"contractAddress"`
	// GasUsed         uint64         `json:"gasUsed" gencodec:"required"`

	// Inclusion information: These fields provide information about the inclusion of the
	// transaction corresponding to this receipt.
	BlockHash        common.Hash `json:"blockHash,omitempty"`
	BlockNumber      *big.Int    `json:"blockNumber,omitempty"`
	TransactionIndex uint        `json:"transactionIndex"`

	// Privacy
	PrivateFrom privacy.PublicKey   `json:"privateFrom"    gencodec:"required"`
	PrivateFor  []privacy.PublicKey `json:"privateFor"    gencodec:"required"`
	Restriction string

	// Private
	CommitmentHash common.Hash `json:"commitmentHash" gencodec:"required"`
	Output         []byte      `json:"output"`
}

// MarshalPrivateReceipt .
func MarshalPrivateReceipt(r map[string]interface{}) (*PrivateReceipt, error) {
	// contractAddress not required
	var contractAddress common.Address
	if v, ok := r["contractAddress"]; ok {
		contractAddress = common.HexToAddress(v.(string))
	}
	// output not required
	var output []byte
	if v, ok := r["output"]; ok {
		output, _ = hexutil.Decode(v.(string))
	}
	// commitmentHash required
	if _, ok := r["commitmentHash"]; !ok {
		return nil, fmt.Errorf("commitmentHash not found")
	}
	commitmentHash := common.HexToHash(r["commitmentHash"].(string))
	// transactionHash required
	if _, ok := r["transactionHash"]; !ok {
		return nil, fmt.Errorf("transactionHash not found")
	}
	transactionHash := common.HexToHash(r["transactionHash"].(string))
	// privateFrom required
	if _, ok := r["privateFrom"]; !ok {
		return nil, fmt.Errorf("privateFrom not found")
	}
	privateFrom, err := privacy.ToPublicKey(r["privateFrom"].(string))
	if err != nil {
		return nil, err
	}
	// privateFor required
	if _, ok := r["privateFor"]; !ok {
		return nil, fmt.Errorf("privateFor not found")
	}
	var privateFor []privacy.PublicKey
	for _, v := range r["privateFor"].([]interface{}) {
		key, err := privacy.ToPublicKey(v.(string))
		if err != nil {
			continue
		}
		privateFor = append(privateFor, key)
	}
	// status not required
	status := uint64(0)
	if v, ok := r["status"]; ok {
		if v.(string) == "0x1" {
			status = uint64(1)
		}
	}
	// logs required
	if _, ok := r["logs"]; !ok {
		return nil, fmt.Errorf("logs not found")
	}
	var logs []*types.Log
	for _, v := range r["logs"].([]interface{}) {
		var log *types.Log
		err := log.UnmarshalJSON(v.([]byte))
		if err != nil {
			continue
		}
		logs = append(logs, log)
	}
	// logsBloom required
	if _, ok := r["logsBloom"]; !ok {
		return nil, fmt.Errorf("logsBloom not found")
	}
	logsBloomString := r["logsBloom"].(string)
	logsBloomBytes, err := hexutil.Decode(logsBloomString)
	if err != nil {
		return nil, fmt.Errorf("failed to Decode %v, err: %v", logsBloomString, err)
	}
	logsBloom := types.BytesToBloom(logsBloomBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to UnmarshalText %v, err: %v", logsBloomString, err)
	}
	// blockHash not required
	var blockHash common.Hash
	if v, ok := r["blockHash"]; ok {
		blockHash = common.HexToHash(v.(string))
	}
	// blockNumber not required
	var blockNumber *big.Int
	if v, ok := r["blockNumber"]; ok {
		i := new(big.Int)
		i.SetString(v.(string), 16)
		blockNumber = i
	}
	// transactionIndex not required
	var transactionIndex uint
	if v, ok := r["transactionIndex"]; ok {
		i := new(big.Int)
		i.SetString(v.(string), 16)
		transactionIndex = uint(i.Uint64())
	}
	return &PrivateReceipt{
		Status:           status,
		Bloom:            logsBloom,
		Logs:             logs,
		TxHash:           transactionHash,
		ContractAddress:  contractAddress,
		BlockHash:        blockHash,
		BlockNumber:      blockNumber,
		TransactionIndex: transactionIndex,
		PrivateFrom:      privateFrom,
		PrivateFor:       privateFor,
		Restriction:      "restricted",
		CommitmentHash:   commitmentHash,
		Output:           output,
	}, nil
}
