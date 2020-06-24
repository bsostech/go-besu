# go-besu
go-besu is a Golang integration tool for Hyperledger Besu, helping developer to build a private raw transaction.

## Privacy
Use privacy of go-besu to get private nonce of an account with one specific privacy group.
- init
    ```go
    priv := privacy.NewPrivacy(rpcClient)
    ```
- get privacy manager public key
    ```go
    privateFrom, _ := privacy.ToPublicKey(privateFromString)
	privateFor1, _ := privacy.ToPublicKey(privateFor1String)
	privateFor2, _ := privacy.ToPublicKey(privateFor2String)
	privateFor := [][]byte{privateFor1, privateFor2}
	participants := []*privacy.PublicKey{&privateFrom, &privateFor1, &privateFor2}
    ```
- find root privacy group
    ```go
    rootPrivacyGroup := priv.FindRootPrivacyGroup(participants)
    ```
- get private nonce
    ```go
    privateNonce, _ := priv.PrivateNonce(fromAddress, rootPrivacyGroup)
    ```

## Types
Use types to go-besu to generate a private raw transaction.
- create a private transaction model
    ```go
    besutx := types.NewContractCreation(privateNonce, nil, gasLimit, big.NewInt(0), data, privateFrom, privateFor)
    besutx := types.NewTransaction(privateNonce, contractAddress, nil, gasLimit, big.NewInt(0), data, privateFrom, privateFor)
    ```
- sign private transaction
    ```go
    besuSignedTx, _ := besutx.SignTx(networkID, privateKey)
    ```
- encode and get private raw transaction
    ```go
    besuRawTxData, _ := rlp.EncodeToBytes(besuSignedTx)
    ```

## Examples
```go
package main

import (
	"context"
	"crypto/ecdsa"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/bsostech/go-besu/privacy"
	"github.com/bsostech/go-besu/types"
)

func main() {
	privateFromString := "A1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo="
	privateFor1String := "Ko2bVqD+nNlNYL5EE7y3IdOnviftjiizpjRt+HTuFBs="
	privateFor2String := "k2zXEin4Ip/qBGlRkJejnGWdP9cjkK+DAvKNW31L2C8="
	privateFrom, _ := privacy.ToPublicKey(privateFromString)
	privateFor1, _ := privacy.ToPublicKey(privateFor1String)
	privateFor2, _ := privacy.ToPublicKey(privateFor2String)
	privateFor := [][]byte{privateFor1, privateFor2}
	participants := []*privacy.PublicKey{&privateFrom, &privateFor1, &privateFor2}

	rpcClient, _ := rpc.Dial("http://localhost:20000")
	ethClient := ethclient.NewClient(rpcClient)
	privateKey, _ := crypto.HexToECDSA("8f2a55949038a9610f50fb23b5883af3b4ecb3c3bb792cbcefbd1542c692be63")
	publicKey := privateKey.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	gasLimit := uint64(3000000)
	networkID, _ := ethClient.NetworkID(context.TODO())

	// get private nonce
	// 1. find private group
	priv := privacy.NewPrivacy(rpcClient)
	rootPrivacyGroup := priv.FindRootPrivacyGroup(participants)
	// 2. get private nonce
	privateNonce, _ := priv.PrivateNonce(fromAddress, rootPrivacyGroup)

	// besutx := types.NewContractCreation(privateNonce, nil, gasLimit, big.NewInt(0), data, privateFrom, privateFor)
	contractAddress := common.HexToAddress("0xaa56458ec6440e480f38be8de3a1abca3a95b7ea")
	data, _ := hexutil.Decode("0x0121b93f0000000000000000000000000000000000000000000000000000000000000002")
	besutx := types.NewTransaction(privateNonce, &contractAddress, nil, gasLimit, big.NewInt(0), data, privateFrom, privateFor)
	besuSignedTx, _ := besutx.SignTx(networkID, privateKey)
	besuRawTxData, _ := rlp.EncodeToBytes(besuSignedTx)
	for i := range besuRawTxData {
	    tmp, _ := rlp.EncodeToBytes(besuSignedTx)
	    tmp = append(tmp[:1], tmp[i:]...)
	    var txSlice []interface{}
	    err = rlp.DecodeBytes(tmp, &txSlice)
	    if err != nil {
	        continue
	    }
	    if len(txSlice) == 12 { // 12 args in private transaction
	        besuRawTxData = tmp
	        break
	    }
	}
	var txHash common.Hash
	rpcClient.CallContext(context.TODO(), &txHash, "eea_sendRawTransaction", hexutil.Encode(besuRawTxData))
	log.Println(txHash.Hex())
}
```
