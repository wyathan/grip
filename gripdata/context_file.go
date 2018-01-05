package gripdata

import (
	"crypto/sha512"

	"github.com/wyathan/grip/gripcrypto"
)

//ContextFile file data pertaining to a context
type ContextFile struct {
	Index       bool
	Snapshot    bool
	DependsOn   [][]byte
	Path        string
	ContextUser string //If a logged in context user submitted this
	NodeID      []byte //Node ID of creating node
	CreatedOn   uint64 //Time stamp
	Context     []byte //Digest of the context this file bellongs to
	DataDepDig  []byte //Digest of Index, Snapshot, DependsOn, Context, and the contents of Path
	Dig         []byte //This record's digest
	Sig         []byte //Signed by NodeID private key
}

func xorBytes(a []byte, b []byte) {
	for c := 0; c < len(a) && c < len(b); c++ {
		a[c] = a[c] ^ b[c]
	}
}

func (a *ContextFile) PushDep(dd []byte) *ContextFile {
	a.DependsOn = append(a.DependsOn, dd)
	return a
}

//Digest ContextFile
func (a *ContextFile) Digest() []byte {
	h := sha512.New()
	gripcrypto.HashBool(h, a.Index)
	gripcrypto.HashBool(h, a.Snapshot)
	depbs := make([]byte, sha512.Size)
	if a.DependsOn == nil {
		gripcrypto.HashUint32(h, 0)
	} else {
		gripcrypto.HashUint32(h, uint32(len(a.DependsOn)))
		//Simply xor the dependency references.  They are
		//already digests, and xoring will eleminate any
		//affect of the order of the dependencies
		for _, d := range a.DependsOn {
			xorBytes(depbs, d)
		}
		gripcrypto.HashBytes(h, depbs)
	}
	gripcrypto.HashBytes(h, a.Context)
	gripcrypto.HashFile(h, a.Path)
	a.DataDepDig = h.Sum(nil)
	h = sha512.New()
	gripcrypto.HashBytes(h, a.DataDepDig)
	gripcrypto.HashString(h, a.ContextUser)
	gripcrypto.HashBytes(h, a.NodeID)
	gripcrypto.HashUint64(h, a.CreatedOn)
	a.Dig = h.Sum(nil)
	return a.Dig
}

//SetSig ContextFile
func (a *ContextFile) SetSig(b []byte) {
	a.Sig = b
}

//GetSig ContextFile
func (a *ContextFile) GetSig() []byte {
	return a.Sig
}

//GetNodeID ContextFile
func (a *ContextFile) GetNodeID() []byte {
	return a.NodeID
}

//GetDig ContextFile
func (a *ContextFile) GetDig() []byte {
	return a.Dig
}

//SetNodeID ContextFile
func (a *ContextFile) SetNodeID(id []byte) {
	a.NodeID = id
}

//checkContextFunctionsImplSign just forces interface check
func checkContextFunctionsImplSign() {
	var f ContextFile
	gripcrypto.Sign(&f, nil)
}
