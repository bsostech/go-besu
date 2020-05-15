package types

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"
)

// PrivateTransaction .
type PrivateTransaction struct {
	Data txdata
}

type txdata struct {
	AccountNonce uint64          `json:"nonce"    gencodec:"required"`
	Price        *big.Int        `json:"gasPrice" gencodec:"required"`
	GasLimit     uint64          `json:"gas"      gencodec:"required"`
	Recipient    *common.Address `json:"to"       rlp:"nil"` // nil means contract creation
	Amount       *big.Int        `json:"value"    gencodec:"required"`
	Payload      []byte          `json:"input"    gencodec:"required"`

	V *big.Int `json:"v" gencodec:"required"`
	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`

	PrivateFrom []byte   `json:"private_from"    gencodec:"required"`
	PrivateFor  [][]byte `json:"private_for"    gencodec:"required"`
	Restriction string
}

// NewContractCreation .
func NewContractCreation(nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, privateFrom []byte, privateFor [][]byte) *PrivateTransaction {
	return newTransaction(nonce, nil, amount, gasLimit, gasPrice, data, privateFrom, privateFor)
}

// SignTx .
func (tx *PrivateTransaction) SignTx(chainID *big.Int, prv *ecdsa.PrivateKey) (*PrivateTransaction, error) {
	h := hash(tx, chainID)
	sig, err := crypto.Sign(h[:], prv)
	if err != nil {
		return nil, err
	}
	return withSignature(tx, sig, chainID)
}

func newTransaction(nonce uint64, to *common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, privateFrom []byte, privateFor [][]byte) *PrivateTransaction {
	if len(data) > 0 {
		data = common.CopyBytes(data)
	}
	d := txdata{
		AccountNonce: nonce,
		Recipient:    to,
		Payload:      data,
		Amount:       new(big.Int),
		GasLimit:     gasLimit,
		Price:        new(big.Int),
		PrivateFrom:  privateFrom,
		PrivateFor:   privateFor,
		Restriction:  "restricted",
		V:            new(big.Int),
		R:            new(big.Int),
		S:            new(big.Int),
	}
	if amount != nil {
		d.Amount.Set(amount)
	}
	if gasPrice != nil {
		d.Price.Set(gasPrice)
	}
	return &PrivateTransaction{Data: d}
}

func hash(tx *PrivateTransaction, chainID *big.Int) common.Hash {
	h := rlpHash([]interface{}{
		tx.Data.AccountNonce,
		tx.Data.Price,
		tx.Data.GasLimit,
		tx.Data.Recipient,
		tx.Data.Amount,
		tx.Data.Payload,
		chainID, uint(0), uint(0),
		tx.Data.PrivateFrom,
		tx.Data.PrivateFor,
		tx.Data.Restriction,
	})
	return h
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewLegacyKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}

func withSignature(tx *PrivateTransaction, sig []byte, chainID *big.Int) (*PrivateTransaction, error) {
	r, s, v, err := signatureValues(tx, sig)
	if err != nil {
		return nil, err
	}
	newV := v.Uint64() + chainID.Uint64()*2 + 8 // KEVIN hack from web3js-eea
	cpy := &PrivateTransaction{Data: tx.Data}
	cpy.Data.R, cpy.Data.S, cpy.Data.V = r, s, new(big.Int).SetUint64(newV)
	return cpy, nil
}

func signatureValues(tx *PrivateTransaction, sig []byte) (r, s, v *big.Int, err error) {
	if len(sig) != crypto.SignatureLength {
		panic(fmt.Sprintf("wrong size for signature: got %d, want %d", len(sig), crypto.SignatureLength))
	}
	r = new(big.Int).SetBytes(sig[:32])
	s = new(big.Int).SetBytes(sig[32:64])
	v = new(big.Int).SetBytes([]byte{sig[64] + 27})
	return r, s, v, nil
}
