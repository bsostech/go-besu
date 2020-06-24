package privacy

import (
	"context"
	"encoding/base64"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"golang.org/x/crypto/sha3"
)

// Privacy .
type Privacy struct {
	client *rpc.Client
}

// Group .
type Group struct {
	ID          string
	Name        string
	Description string
	Type        string
	Members     []*PublicKey
}

// PublicKey .
type PublicKey []byte

// NewPrivacy .
func NewPrivacy(c *rpc.Client) *Privacy {
	return &Privacy{
		client: c,
	}
}

// PrivateNonceByParticipants .
func (p *Privacy) PrivateNonceByParticipants(account common.Address, participants []*PublicKey) (uint64, error) {
	rootGroup := p.FindRootPrivacyGroup(participants)
	return p.PrivateNonce(account, rootGroup)
}

// FindRootPrivacyGroup .
func (p *Privacy) FindRootPrivacyGroup(participants []*PublicKey) *Group {
	sortParticipants := p.sort(participants)
	hash := rlpHash(sortParticipants)
	return &Group{
		ID: base64.StdEncoding.EncodeToString(hash.Bytes()),
	}
}

// PrivateNonce .
func (p *Privacy) PrivateNonce(account common.Address, privacyGroup *Group) (uint64, error) {
	var getTransactionCountRsp interface{}
	err := p.client.CallContext(context.TODO(), &getTransactionCountRsp, "priv_getTransactionCount", account.Hex(), privacyGroup.ID)
	if err != nil {
		return 0, err
	}
	nonce, err := hexutil.DecodeUint64(getTransactionCountRsp.(string))
	if err != nil {
		return 0, err
	}
	return nonce, nil
}

// FindPrivacyGroup .
func (p *Privacy) FindPrivacyGroup(participants []*PublicKey) (*Group, error) {
	publicKeysString := make([]string, len(participants))
	for i := range participants {
		publicKeysString[i] = participants[i].ToString()
	}
	var findPrivacyGroupRsp []map[string]interface{}
	err := p.client.CallContext(context.TODO(), &findPrivacyGroupRsp, "priv_findPrivacyGroup", participants)
	if err != nil {
		return nil, err
	}
	var privacyGroup Group
	if len(findPrivacyGroupRsp) == 0 {
		return nil, nil
	}
	ms := findPrivacyGroupRsp[0]["members"].([]interface{})
	var members []*PublicKey
	for _, v := range ms {
		m, err := ToPublicKey(v.(string))
		if err != nil {
			continue
		}
		members = append(members, &m)
	}
	privacyGroup.ID = findPrivacyGroupRsp[0]["privacyGroupId"].(string)
	privacyGroup.Name = findPrivacyGroupRsp[0]["name"].(string)
	privacyGroup.Description = findPrivacyGroupRsp[0]["description"].(string)
	privacyGroup.Type = findPrivacyGroupRsp[0]["type"].(string)
	privacyGroup.Members = members
	return &privacyGroup, nil
}

// CreatePrivacyGroup .
func (p *Privacy) CreatePrivacyGroup(members []*PublicKey, name string) (*Group, error) {
	args := getCreatePrivacyGroupArgs(members, name)
	var createPrivacyGroupRsp interface{}
	err := p.client.CallContext(context.TODO(), &createPrivacyGroupRsp, "priv_createPrivacyGroup", args)
	if err != nil {
		return nil, err
	}
	return &Group{
		ID:      createPrivacyGroupRsp.(string),
		Name:    name,
		Members: members,
	}, nil
}

// ToPublicKey .
func ToPublicKey(key string) (PublicKey, error) {
	return base64.StdEncoding.DecodeString(key)
}

// ToString .
func (pub PublicKey) ToString() string {
	return base64.StdEncoding.EncodeToString(pub)
}

// Hash .
func (pub PublicKey) Hash() int {
	result := int(1)
	for _, v := range pub {
		result = int(int32((31*result + int((int32(v)<<24)>>24)) & 0xffffffff))
	}
	return result
}

func getCreatePrivacyGroupArgs(publicKeys []*PublicKey, name string) map[string]interface{} {
	publicKeysString := make([]string, len(publicKeys))
	for i := range publicKeys {
		publicKeysString[i] = publicKeys[i].ToString()
	}
	result := make(map[string]interface{})
	result["addresses"] = publicKeysString
	result["name"] = name
	return result
}

// hack from web3js-eea src/privacyGroup.js
func (p *Privacy) sort(participants []*PublicKey) []*PublicKey {
	hashMap := make(map[int]*PublicKey)
	for i := range participants {
		hashMap[participants[i].Hash()] = participants[i]
	}
	var keys []int
	for k := range hashMap {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	var output []*PublicKey
	for _, v := range keys {
		output = append(output, hashMap[v])
	}
	return output
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewLegacyKeccak256()
	err := rlp.Encode(hw, x)
	if err != nil {
		return common.Hash{}
	}
	hw.Sum(h[:0])
	return h
}
