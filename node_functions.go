package grip

import (
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"log"
	"reflect"
	"time"

	"encoding/base64"

	"github.com/wyathan/grip/gripcrypto"
	"github.com/wyathan/grip/gripdata"
)

func ListShareWithMe() []gripdata.ShareNodeInfo {
	return nil
}

//NewShareNode states that this node's Node data should be
//shared with another node
func NewShareNode(n *gripdata.ShareNodeInfo, db NodeNetAccountdb) error {
	myn, mypr := db.GetPrivateNodeData()
	id := myn.ID
	n.SetNodeID(id)

	//See if the target node already has an account
	createAutoAccount(mypr, n.TargetNodeID, db)

	err := SignNodeSig(n, db)
	if err != nil {
		return err
	}
	err = db.StoreShareNodeInfo(n)
	if err != nil {
		return err
	}
	nd := db.GetNode(id)
	if nd == nil {
		return errors.New("This node is not found")
	}
	err = SendAllToShareWith(nd, db)
	if err != nil {
		return err
	}
	err = SendAllToShareWith(n, db)
	if err != nil {
		return err
	}
	err = SendAllSharesToNew(id, n.TargetNodeID, db)
	if err != nil {
		return err
	}
	return nil
}

//NewUseShareNodeKey states that this node's Node data should be
//shared with another node
func NewUseShareNodeKey(n *gripdata.UseShareNodeKey, db NodeNetdb) error {
	myn, _ := db.GetPrivateNodeData()
	id := myn.ID
	n.SetNodeID(id)
	if n.Key == "" {
		return errors.New("Key cannot be empty")
	}
	err := SignNodeSig(n, db)
	if err != nil {
		return err
	}
	err = db.StoreUseShareNodeKey(n)
	if err != nil {
		return err
	}
	nd := db.GetNode(id)
	if nd == nil {
		return errors.New("This node is not found")
	}
	err = CreateNewSend(n, n.TargetID, db)
	if err != nil {
		return err
	}
	return nil
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

//SendAllToAllKey sends new data to all nodes on share list
func SendAllToAllKey(s gripcrypto.SignInf, k string, db NodeNetdb) error {
	shl := FindAllToShareWithFromKey(k, db)
	for _, sh := range shl {
		err := CreateNewSend(s, sh, db)
		if err != nil {
			return err
		}
	}
	return nil
}

//FindAllToShareWithFromKey get all nodes to send data to based on key
func FindAllToShareWithFromKey(k string, db NodeNetdb) [][]byte {
	var s [][]byte
	kl := db.ListShareNodeKey(k)
	for _, sn := range kl {
		s = append(s, sn.NodeID)
		s = append(s, FindAllToShareWith(sn.NodeID, db)...)
	}
	return s
}

//FindAllToShareWith get all the nodes to share updates about a node
func FindAllToShareWith(id []byte, db NodeNetdb) [][]byte {
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
func SendAllSharesToNew(from []byte, to []byte, db NodeNetdb) error {
	shl := db.ListShareNodeInfo(from)
	km := make(map[string]bool)
	for _, sr := range shl {
		n := db.GetNode(sr.NodeID)
		nd := db.GetNode(sr.TargetNodeID)
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

//SendAllToShareWith sends new data to all nodes on share list
func SendAllToShareWith(s gripcrypto.SignInf, db NodeNetdb) error {
	shl := FindAllToShareWith(s.GetNodeID(), db)
	for _, sh := range shl {
		err := CreateNewSend(s, sh, db)
		if err != nil {
			return err
		}
	}
	return nil
}

//CreateNewSend indicates this data should be sent to this node
func CreateNewSend(v gripcrypto.SignInf, sendto []byte, db Netdb) error {
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

//SignNodeSig sign a new sign interface
func SignNodeSig(s gripcrypto.SignInf, db Nodedb) error {
	_, privdata := db.GetPrivateNodeData()
	s.SetNodeID(privdata.ID)
	err := gripcrypto.Sign(s, privdata.PrivateKey)
	if err != nil {
		return err
	}
	return nil
}

//VerifyNodeSig verify a node signed a sign interface
func VerifyNodeSig(s gripcrypto.SignInf, db Nodedb) error {
	nodeid := s.GetNodeID()
	if nodeid == nil {
		return errors.New("NodeID was not set")
	}
	nodedata := db.GetNode(nodeid)
	if nodedata == nil {
		return errors.New("Node was not found in db")
	}
	if !gripcrypto.Verify(s, nodedata.PublicKey) {
		return errors.New("Signature was not valid")
	}
	return nil
}
