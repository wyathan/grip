package gripdata

import (
	"crypto/sha512"

	"github.com/wyathan/grip/gripcrypto"
)

//AssociateNodeAccountKey this is pushed to a Node where the user
//may have an account they would like this new node to be assocated with.
//Once the node is assocated with an account on the other node, certain
//permissions may be granted to the node based on the account on that
//node.  This information should never be shared beyond the NodeID and
//TargetNodeID nodes.
type AssociateNodeAccountKey struct {
	NodeID       []byte //Subject node to assocate with account
	TargetNodeID []byte //The only node that should get this key where the account is
	Key          string //The magic key
	Dig          []byte //This record's digest
	Sig          []byte //Signed by NodeID private key
}

//Digest AssociateNodeAccountKey
func (a *AssociateNodeAccountKey) Digest() []byte {
	h := sha512.New()
	gripcrypto.HashString(h, a.Key)
	gripcrypto.HashBytes(h, a.NodeID)
	gripcrypto.HashBytes(h, a.TargetNodeID)
	a.Dig = h.Sum(nil)
	return a.Dig
}

//SetSig AssociateNodeAccountKey
func (a *AssociateNodeAccountKey) SetSig(b []byte) {
	a.Sig = b
}

//GetSig AssociateNodeAccountKey
func (a *AssociateNodeAccountKey) GetSig() []byte {
	return a.Sig
}

//GetNodeID AssociateNodeAccountKey
func (a *AssociateNodeAccountKey) GetNodeID() []byte {
	return a.NodeID
}

//GetDig AssociateNodeAccountKey
func (a *AssociateNodeAccountKey) GetDig() []byte {
	return a.Dig
}

//SetNodeID AssociateNodeAccountKey
func (a *AssociateNodeAccountKey) SetNodeID(id []byte) {
	a.NodeID = id
}

func checkAssociateNodeAccountKey() {
	var a AssociateNodeAccountKey
	gripcrypto.Sign(&a, nil)
}
