package griptests

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/wyathan/grip/gripcrypto"
	"github.com/wyathan/grip/gripdata"
)

func TestNodeGeneratedAccountSig(t *testing.T) {
	prv, pub, err := gripcrypto.GenerateECDSAKeyPair()
	if err != nil {
		t.Error("Error generating key")
	}
	var n gripdata.NodeGeneratedAccount
	n.AccountID = "AAAAAAAAA00000"
	n.MetaData = "Metadata"
	n.NodeID = make([]byte, 10)
	rand.Read(n.NodeID)
	n.TargetNodeID = make([]byte, 12)
	rand.Read(n.TargetNodeID)

	err = gripcrypto.Sign(&n, prv)
	if err != nil {
		t.Errorf("Failed to sign: %s", err)
	}
	if len(n.Sig) <= 0 {
		t.Error("Invalid signature")
	}
	var sig = n.GetSig()
	if 0 != bytes.Compare(sig, n.Sig) {
		t.Error("Invalid GetSig")
	}

	if !gripcrypto.Verify(&n, pub) {
		t.Error("Failed to verify sig")
	}

	n.AccountID = "AAAAAAAAA00001"
	if gripcrypto.Verify(&n, pub) {
		t.Error("Failed to verify fail")
	}

}

func TestNodeGeneratedAccountDigest(t *testing.T) {
	var n, m gripdata.NodeGeneratedAccount
	n.AccountID = "AAAAAAAAA00000"
	n.MetaData = "Metadata"
	n.NodeID = make([]byte, 10)
	rand.Read(n.NodeID)
	n.TargetNodeID = make([]byte, 12)
	rand.Read(n.TargetNodeID)
	n.Digest()
	if n.Dig == nil {
		t.Error("Digest is nil")
	}
	s := hex.EncodeToString(n.Dig)
	fmt.Println(s)

	m.AccountID = "AAAAAAAAA00000"
	m.MetaData = "Metadata"
	m.NodeID = make([]byte, 10)
	m.NodeID = n.NodeID
	m.TargetNodeID = n.TargetNodeID
	m.Digest()

	if !bytes.Equal(n.Dig, m.Dig) {
		t.Error("Digests don't match")
	}

	m.AccountID = "AAAAAAAAA00001"
	m.Digest()

	if bytes.Equal(n.Dig, m.Dig) {
		t.Error("Digests match")
	}

}

func TestContextFileDigest(t *testing.T) {
	var n, m gripdata.ContextFile
	n.Context = make([]byte, 10)
	rand.Read(n.Context)
	n.NodeID = make([]byte, 12)
	rand.Read(n.NodeID)
	n.ContextUser = "brownie"
	n.CreatedOn = 12345
	n.Index = true
	n.Snapshot = false
	n.DependsOn = make([][]byte, 5)
	for c := 0; c < 5; c++ {
		n.DependsOn[c] = make([]byte, 17)
		rand.Read(n.DependsOn[c])
	}
	b := []byte("This is a test file for checking ContextFile digest")
	ioutil.WriteFile("testfile", b, 0x777)
	n.SetPath("testfile")
	n.Digest()
	if n.Dig == nil {
		t.Error("Digest is nil")
	}
	s := hex.EncodeToString(n.Dig)
	fmt.Print("ContextFile dig: ")
	fmt.Println(s)

	m = n

	m.Digest()

	if !bytes.Equal(n.Dig, m.Dig) {
		t.Error("Digests don't match")
	}

	b = []byte("Th1s is a test file for checking ContextFile digest")
	ioutil.WriteFile("testfile2", b, 0x777)
	m.SetPath("testfile2")

	m.Digest()

	if bytes.Equal(n.Dig, m.Dig) {
		t.Error("Digests match")
	}

	m = n
	m.DependsOn = make([][]byte, len(n.DependsOn))
	for c := 0; c < len(n.DependsOn); c++ {
		m.DependsOn[c] = make([]byte, len(n.DependsOn[c]))
		copy(m.DependsOn[c], n.DependsOn[c])
	}

	m.Digest()

	if !bytes.Equal(n.Dig, m.Dig) {
		t.Error("Digests don't match")
	}

	m.DependsOn[1][5] = m.DependsOn[1][5] + 9

	m.Digest()

	if bytes.Equal(n.Dig, m.Dig) {
		t.Error("Digests match")
	}

}
