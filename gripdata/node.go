package gripdata

import (
	"crypto/sha512"

	"github.com/wyathan/grip/gripcrypto"
)

//Node is a computer that can participate
//in contexts
type Node struct {
	ID          []byte //Unique id for the node (digest of public key)
	Name        string
	PublicKey   []byte
	URL         string //where we can find this node
	Connectable bool   //Can we directly connect to it
	MetaData    string //Generic metadata string
	Dig         []byte //Digest of all data
	Sig         []byte //Signed private key of this node
}

//NodeEphemera is local data kept about this node, it is used to query
//for nodes we wish to initiate connections
type NodeEphemera struct {
	ID                []byte //Unique id for the node
	Connectable       bool   //Can we directly connect to it
	LastConnAttempt   uint64 //Last time we attempted to connect this node
	LastConnection    uint64 //Last time we successfully connected to this node
	LastConnReceived  uint64 //Last time this node connected to us
	NextAttempt       uint64 //The next time we should attempt to connect to this node
	ConnectionPending bool
	Connected         bool //Are we currently connected
}

//Digest Node
func (a *Node) Digest() []byte {
	h := sha512.New()
	gripcrypto.HashBool(h, a.Connectable)
	gripcrypto.HashBytes(h, a.ID)
	gripcrypto.HashString(h, a.MetaData)
	gripcrypto.HashString(h, a.Name)
	gripcrypto.HashBytes(h, a.PublicKey)
	gripcrypto.HashString(h, a.URL)
	a.Dig = h.Sum(nil)
	return a.Dig
}

//SetSig Node
func (a *Node) SetSig(b []byte) {
	a.Sig = b
}

//GetSig Node
func (a *Node) GetSig() []byte {
	return a.Sig
}

//GetNodeID Node
func (a *Node) GetNodeID() []byte {
	return a.ID
}

//GetDig Node
func (a *Node) GetDig() []byte {
	return a.Dig
}

//SetNodeID Node
func (a *Node) SetNodeID(id []byte) {
	a.ID = id
}

//checkNode just forces interface check
func checkNode() {
	var f Node
	gripcrypto.Sign(&f, nil)
}
