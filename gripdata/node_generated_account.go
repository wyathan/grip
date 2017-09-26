package gripdata

import (
	"crypto/sha512"

	"github.com/wyathan/grip/gripcrypto"
)

//NodeGeneratedAccount is an account associated generated
//for one node to assocated it with an account on another node.
//This data should only be shared between the NodeID and TargetNodeID
//nodes.
type NodeGeneratedAccount struct {
	NodeID       []byte //The id of the node that belongs to the account
	TargetNodeID []byte //The id of the node that belongs to the account
	AccountID    string //The account id
	MetaData     string //Generic metadata string
	Dig          []byte //Digest of all data
	Sig          []byte //Signed private key of this node
}

//Digest NodeGeneratedAccount
func (a *NodeGeneratedAccount) Digest() []byte {
	h := sha512.New()
	gripcrypto.HashBytes(h, a.NodeID)
	gripcrypto.HashBytes(h, a.TargetNodeID)
	gripcrypto.HashString(h, a.AccountID)
	gripcrypto.HashString(h, a.MetaData)
	a.Dig = h.Sum(nil)
	return a.Dig
}

//SetSig NodeGeneratedAccount
func (a *NodeGeneratedAccount) SetSig(b []byte) {
	a.Sig = b
}

//GetSig NodeGeneratedAccount
func (a *NodeGeneratedAccount) GetSig() []byte {
	return a.Sig
}

//GetNodeID NodeGeneratedAccount
func (a *NodeGeneratedAccount) GetNodeID() []byte {
	return a.NodeID
}

//GetDig NodeGeneratedAccount
func (a *NodeGeneratedAccount) GetDig() []byte {
	return a.Dig
}

//SetNodeID NodeGeneratedAccount
func (a *NodeGeneratedAccount) SetNodeID(id []byte) {
	a.NodeID = id
}

//checkNodeGeneratedAccount just forces interface check
func checkNodeGeneratedAccount() {
	var f NodeGeneratedAccount
	gripcrypto.Sign(&f, nil)
}
