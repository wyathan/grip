package grip

import (
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"log"
	"reflect"
	"time"

	"encoding/base64"

	"github.com/wyathan/grip/gripcrypto"
	"github.com/wyathan/grip/gripdata"
	"github.com/wyathan/grip/griperrors"
)

type SNProc func(s *gripdata.ShareNodeInfo, db DB) (bool, error)

func ShareNodeProc(np SNProc) SProc {
	return func(s gripcrypto.SignInf, db DB) (bool, error) {
		switch n := s.(type) {
		case *gripdata.ShareNodeInfo:
			return np(n, db)
		}
		return false, griperrors.NotShareNode
	}
}

type UKProc func(s *gripdata.UseShareNodeKey, db DB) (bool, error)

func UseKeyProc(np UKProc) SProc {
	return func(s gripcrypto.SignInf, db DB) (bool, error) {
		switch n := s.(type) {
		case *gripdata.UseShareNodeKey:
			return np(n, db)
		}
		return false, griperrors.NotShareNode
	}
}

func StoreShareNodeInfo(v *gripdata.ShareNodeInfo, db DB) (bool, error) {
	return true, db.StoreShareNodeInfo(v)
}

func SendAllShareNodeInfo(v *gripdata.ShareNodeInfo, db DB) (bool, error) {
	return true, SendAllSharesToNew(v.NodeID, v.TargetNodeID, db)
}

//NewShareNode states that this node's Node data should be
//shared with another node
func NewShareNode(n *gripdata.ShareNodeInfo, db DB) error {
	myn, mypr := db.GetPrivateNodeData()
	id := myn.ID
	n.SetNodeID(id)

	createAutoAccount(mypr, n.TargetNodeID, db)
	var v SProcChain
	v.Push(SignNodeSig)
	v.Push(ShareNodeProc(StoreShareNodeInfo))
	v.PushV(SendAllToShareWithMe, myn)
	v.Push(SendAllToShareWithMe)
	v.Push(ShareNodeProc(SendAllShareNodeInfo))

	_, err := v.F(n, db)
	return err
}

func StoreUseShareNodeKey(v *gripdata.UseShareNodeKey, db DB) (bool, error) {
	return true, db.StoreUseShareNodeKey(v)
}

func SendAllUseShareNodeKey(v *gripdata.UseShareNodeKey, db DB) (bool, error) {
	return true, CreateNewSend(v, v.TargetID, db)
}

//NewUseShareNodeKey states that this node's Node data should be
//shared with another node
func NewUseShareNodeKey(n *gripdata.UseShareNodeKey, db DB) error {
	myn, _ := db.GetPrivateNodeData()
	id := myn.ID
	n.SetNodeID(id)
	if n.Key == "" {
		return griperrors.ShareNodeKeyEmpty
	}

	var v SProcChain
	v.Push(SignNodeSig)
	v.Push(UseKeyProc(StoreUseShareNodeKey))
	v.Push(UseKeyProc(SendAllUseShareNodeKey))

	_, err := v.F(n, db)
	return err
}

func createAutoAccount(mypr *gripdata.MyNodePrivateData, target []byte, db Accountdb) {
	na := db.GetNodeAccount(target)
	if na == nil && mypr.AutoCreateShareAccount {
		var a gripdata.Account
		b := make([]byte, 32)
		rand.Reader.Read(b)
		a.AccountID = base64.StdEncoding.EncodeToString(b)
		log.Printf("Automatically creating account %s\n", a.AccountID)
		a.AllowCacheMode = mypr.AutoAccountAllowCacheMode
		a.AllowContextNode = mypr.AutoAccountAllowContextNode
		a.AllowContextSource = mypr.AutoAccountAllowContextSource
		a.AllowFullRepo = mypr.AutoAccountAllowFullRepo
		a.AllowNodeAcocuntKey = mypr.AutoAccountAllowNodeAcocuntKey
		a.MaxDiskSpace = mypr.AutoAccountMaxDiskSpace
		a.MaxNodes = mypr.AutoAccountMaxNodes
		a.Enabled = true
		err := db.StoreAccount(&a)
		if err != nil {
			log.Printf("Failed to auto-create account: %s\n", err)
		}
		var na gripdata.NodeAccount
		na.AccountID = a.AccountID
		na.Enabled = true
		na.NodeID = target
		err = db.StoreNodeAccount(&na)
		if err != nil {
			log.Printf("Failed to auto-create account: %s\n", err)
		}
	}
}

//FindAllToShareWithFromKey get all the nodes that we should send
//new data to based on a key.
func FindAllToShareWithFromKey(k string, db DB) [][]byte {
	var s [][]byte
	kl := db.ListShareNodeKey(k)
	for _, sn := range kl {
		s = append(s, sn.NodeID)
		s = append(s, FindAllToShareWith(sn.NodeID, db)...)
	}
	return s
}

//FindAllToShareWith get the targets ids from all the ShareNodeInfo's
//created by id.  Get all the node ids that have created UseShareNodeKey
//that match any of the keys from the ShareNodeInfo created by id.
func FindAllToShareWith(id []byte, db DB) [][]byte {
	sl := db.ListShareNodeInfo(id)
	km := make(map[string]bool)
	bm := make(map[string][]byte)
	for _, sr := range sl {
		bm[base64.StdEncoding.EncodeToString(sr.TargetNodeID)] = sr.TargetNodeID
		if sr.Key != "" {
			km[sr.Key] = true
		}
	}
	for k := range km {
		//Find all nodes that used key
		skl := db.ListUseShareNodeKey(k)
		for _, sk := range skl {
			bm[base64.StdEncoding.EncodeToString(sk.NodeID)] = sk.NodeID
		}
	}
	var s [][]byte
	for _, v := range bm {
		s = append(s, v)
	}
	return s
}

//SendAllSharesToNew sends all data to a new share node
func SendAllSharesToNew(from []byte, to []byte, db DB) error {
	shl := db.ListShareNodeInfo(from)
	km := make(map[string]bool)
	for _, sr := range shl {
		n := db.GetNode(sr.NodeID)
		nd := db.GetNode(sr.TargetNodeID)
		log.Printf("Sending sharenodeinfo not nil: source node %t, target %t", (nd != nil), (n != nil))
		if nd != nil && n != nil {
			//FIXME: You'll send the same node data more than once. :/
			//Although after the first is sent it should remove all matcing
			//SendData from the db
			err := CreateNewSend(n, to, db)
			if err != nil {
				return err
			}
			err = CreateNewSend(nd, to, db)
			if err != nil {
				return err
			}
			err = CreateNewSend(&sr, to, db)
			if err != nil {
				return err
			}
			if sr.Key != "" {
				km[sr.Key] = true
			}
		}
	}
	for k := range km {
		//Find all nodes that used key
		skl := db.ListUseShareNodeKey(k)
		for _, sk := range skl {
			nd := db.GetNode(sk.NodeID)
			err := CreateNewSend(nd, to, db)
			if err != nil {
				return err
			}
			err = CreateNewSend(&sk, to, db)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//SendAllToShareWithMe sends new data to all nodes the source has shared with
func SendAllToShareWithMe(s gripcrypto.SignInf, db DB) (bool, error) {
	shl := FindAllToShareWith(s.GetNodeID(), db)
	for _, sh := range shl {
		err := CreateNewSend(s, sh, db)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

//CreateNewSend indicates this data should be sent to this node
func CreateNewSend(v gripcrypto.SignInf, sendto []byte, db DB) error {
	me, _ := db.GetPrivateNodeData()
	if bytes.Equal(sendto, me.ID) {
		return nil
	}
	var s gripdata.SendData
	s.Dig = v.GetDig()
	s.TargetID = sendto
	s.TypeName = reflect.TypeOf(v).String()
	s.Timestamp = uint64(time.Now().UnixNano())
	err := db.StoreSendData(&s)
	if err != nil {
		return err
	}
	return nil
}

//CreateNewNode creates a new key pair for the node, saves
//the private node data, generates the id, and signs it
func CreateNewNode(pr *gripdata.MyNodePrivateData, n *gripdata.Node, db Nodedb) error {
	if pr.PrivateKey == nil || n.PublicKey == nil {
		prv, pub, err := gripcrypto.GenerateECDSAKeyPair()
		if err != nil {
			return err
		}
		pr.PrivateKey = prv
		n.PublicKey = pub
		h := sha512.New()
		gripcrypto.HashBytes(h, n.PublicKey)
		n.ID = h.Sum(nil)
		pr.ID = n.ID
	}
	err := gripcrypto.Sign(n, pr.PrivateKey)
	if err != nil {
		return err
	}
	err = db.StoreMyPrivateNodeData(n, pr)
	if err != nil {
		return err
	}
	return nil
}

//AssociateNodeAccoutKey associates this node with an account
//on another node
func AssociateNodeAccoutKey(key string, tnid []byte, db DB) error {
	var nac gripdata.AssociateNodeAccountKey
	nac.Key = key
	nac.TargetNodeID = tnid
	_, err := SignNodeSig(&nac, db)
	if err != nil {
		return err
	}
	err = db.StoreAssociateNodeAccountKey(&nac)
	if err != nil {
		return err
	}
	err = CreateNewSend(&nac, tnid, db)
	return err
}

//SignNodeSig sign a new sign interface
func SignNodeSig(s gripcrypto.SignInf, db DB) (bool, error) {
	_, privdata := (Nodedb(db)).GetPrivateNodeData()
	s.SetNodeID(privdata.ID)
	err := gripcrypto.Sign(s, privdata.PrivateKey)
	if err != nil {
		return false, err
	}
	return true, nil
}

//VerifyNodeSig verify a node signed a sign interface
func VerifyNodeSig(s gripcrypto.SignInf, db DB) (bool, error) {
	nodeid := s.GetNodeID()
	if nodeid == nil {
		return false, errors.New("NodeID was not set")
	}
	nodedata := db.GetNode(nodeid)
	if nodedata == nil {
		return false, errors.New("Node was not found in db")
	}
	if !gripcrypto.Verify(s, nodedata.PublicKey) {
		return false, errors.New("Signature was not valid")
	}
	return true, nil
}
