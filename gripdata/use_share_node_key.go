package gripdata

import (
	"crypto/sha512"

	"github.com/wyathan/grip/gripcrypto"
)

//UseShareNodeKey nodes can present this to
//other nodes to get into about another node
type UseShareNodeKey struct {
	NodeID   []byte
	TargetID []byte //The target node that knows of the other node
	Key      string
	Dig      []byte
	Sig      []byte
}

//Digest UseShareNodeKey
func (a *UseShareNodeKey) Digest() []byte {
	h := sha512.New()
	gripcrypto.HashBytes(h, a.NodeID)
	gripcrypto.HashBytes(h, a.TargetID)
	gripcrypto.HashString(h, a.Key)
	a.Dig = h.Sum(nil)
	return a.Dig
}

//SetSig UseShareNodeKey
func (a *UseShareNodeKey) SetSig(b []byte) {
	a.Sig = b
}

//GetSig UseShareNodeKey
func (a *UseShareNodeKey) GetSig() []byte {
	return a.Sig
}

//GetNodeID UseShareNodeKey
func (a *UseShareNodeKey) GetNodeID() []byte {
	return a.NodeID
}

//GetDig UseShareNodeKey
func (a *UseShareNodeKey) GetDig() []byte {
	return a.Dig
}

//SetNodeID UseShareNodeKey
func (a *UseShareNodeKey) SetNodeID(id []byte) {
	a.NodeID = id
}

//checkShareNodeKey just forces interface check
func checkUseShareNodeKey() {
	var f UseShareNodeKey
	gripcrypto.Sign(&f, nil)
}
