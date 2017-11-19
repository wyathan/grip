package grip

import (
	"bytes"
	"crypto/sha512"
	"errors"
	"log"
	"time"

	"github.com/wyathan/grip/gripcrypto"
	"github.com/wyathan/grip/gripdata"
)

//IncomingNode process incoming node data
//Forward to all nodes that it has shared with
func IncomingNode(n *gripdata.Node, db NodeNetAccountdb) error {
	_, pr := db.GetPrivateNodeData()
	if bytes.Equal(pr.ID, n.ID) {
		//I created this.  Ignore it
		return nil
	}
	err := VerifyNode(n)
	if err != nil {
		return err
	}
	err = db.StoreNode(n)
	if err != nil {
		return err
	}
	ne := db.GetNodeEphemera(n.ID)
	if ne == nil {
		var neph gripdata.NodeEphemera
		neph.ID = n.ID
		ne = &neph
	}
	ne.Connectable = n.Connectable
	err = db.StoreNodeEphemera(ne)
	if err != nil {
		return err
	}
	createAutoAccount(pr, n.ID, db)
	if !IsAccountEnabled(n, db) {
		return errors.New("Account is not enabled")
	}
	//Find all ShareNodeInfo for this node
	err = SendAllToShareWithMe(n, db)
	if err != nil {
		return err
	}
	return nil
}

//IncomingShareNode process incoming ShareNodeInfo
func IncomingShareNode(s *gripdata.ShareNodeInfo, db NodeNetAccountdb) error {
	_, pr := db.GetPrivateNodeData()
	if bytes.Equal(pr.ID, s.NodeID) {
		//I created this.  Ignore it
		return nil
	}
	tn := s.TargetNodeID
	if tn == nil {
		return errors.New("TargetNodeID cannot be nil")
	}
	err := VerifyNodeSig(s, db)
	if err != nil {
		return err
	}
	createAutoAccount(pr, s.NodeID, db)
	if !IsAccountEnabled(s, db) {
		return errors.New("Node account is not enabled")
	}
	err = db.StoreShareNodeInfo(s)
	if err != nil {
		return err
	}
	if bytes.Equal(pr.ID, tn) {
		log.Println("Sharing node data with me")
		//Yippy skippy.  They want to share their data with me
		//Do we want to auto-reciprocate?
		//NOTE: This is dangerous.  A public node should
		//Never have this on or anyone could connect share their
		//info and get information about all other nodes.
		//It is also broken in that we can get into a loop
		//between two nodes that both have this on
		//if pr.AutoShareNodeInfo {
		//	log.Println("Autoshare is enabled")
		//	var ns gripdata.ShareNodeInfo
		//	ns.TargetNodeID = s.NodeID
		//	ns.MetaData = pr.AutoShareMetaData
		//	err := NewShareNode(&ns, db)
		//	if err != nil {
		//		log.Printf("Failed to create autoshare: %s\n", err)
		//	}
		//}
		return nil
	}
	//Another node wants to share their node data with someone else
	//Should we check if they have shared with us first?  They already
	//did, but could be a bug.

	//First send their node data, so that they can validate the signature
	//of the ShareNodeInfo
	n := db.GetNode(s.NodeID)
	err = SendAllToShareWithMe(n, db)
	if err != nil {
		return err
	}

	//Send the target node data
	gn := db.GetNode(tn)
	if gn == nil {
		return errors.New("I don't know the target node for this ShareNodeInfo")
	}
	err = SendAllToShareWith(gn, n.ID, db)
	if err != nil {
		return err
	}

	//Now send ShareNodeInfo.
	err = SendAllToShareWithMe(s, db)
	if err != nil {
		return err
	}

	//Now send all the share data to the new node
	err = SendAllSharesToNew(s.NodeID, s.TargetNodeID, db)
	if err != nil {
		return err
	}
	return nil
}

//IncomingUseShareNodeKey processes an incoming UseShareNodeKey
func IncomingUseShareNodeKey(k *gripdata.UseShareNodeKey, db NodeNetAccountdb) error {
	if k.Key == "" {
		return errors.New("Key cannot be empty")
	}
	_, pr := db.GetPrivateNodeData()
	if bytes.Equal(pr.ID, k.NodeID) {
		//I created this.  Ignore it
		return nil
	}
	err := VerifyNodeSig(k, db)
	if err != nil {
		return err
	}
	createAutoAccount(pr, k.NodeID, db)
	if !IsAccountEnabled(k, db) {
		return errors.New("Node account is not enabled")
	}
	err = db.StoreUseShareNodeKey(k)
	if err != nil {
		return err
	}

	n := db.GetNode(k.NodeID)
	if n == nil {
		return errors.New("Could not find node")
	}
	err = SendAllToAllKey(n, k.Key, db)
	if err != nil {
		return err
	}
	err = SendAllToAllKey(k, k.Key, db)
	if err != nil {
		return err
	}

	kl := db.ListShareNodeKey(k.Key)
	for _, v := range kl {
		SendAllSharesToNew(v.NodeID, k.NodeID, db)
	}

	return nil
}

//VerifyNode verify new node data
func VerifyNode(n *gripdata.Node) error {
	if n.PublicKey == nil || n.ID == nil {
		return errors.New("Public key and ID must be specified")
	}
	h := sha512.New()
	gripcrypto.HashBytes(h, n.PublicKey)
	tid := h.Sum(nil)
	if !bytes.Equal(n.ID, tid) {
		return errors.New("ID does not match public key")
	}
	if !gripcrypto.Verify(n, n.PublicKey) {
		return errors.New("Failed to verify signature")
	}
	return nil
}

//GetNodeAccount see if this node id has an account
func GetNodeAccount(id []byte, db Accountdb) *gripdata.Account {
	na := db.GetNodeAccount(id)
	if na == nil {
		return nil
	}
	if !na.Enabled {
		return nil
	}
	return db.GetAccount(na.AccountID)
}

//IsIDAccountEnabled see if this node id has account enabled
func IsIDAccountEnabled(id []byte, db Accountdb) bool {
	a := GetNodeAccount(id, db)
	if a == nil {
		return false
	}
	return a.Enabled
}

//IsAccountEnabled see if this signInf has account enabled
func IsAccountEnabled(s gripcrypto.SignInf, db Accountdb) bool {
	return IsIDAccountEnabled(s.GetNodeID(), db)
}

//IncomingNodeAccountKey process incoming ShareNodeInfoKeys
func IncomingNodeAccountKey(s *gripdata.AssociateNodeAccountKey, db NodeAccountdb) error {
	err := VerifyNodeSig(s, db)
	if err != nil {
		return err
	}
	ak := db.GetNodeAccountKey(s.Key)
	if ak == nil {
		return errors.New("Node account key not found")
	}
	if ak.Used {
		return errors.New("Onetime node account key has been used")
	}
	nw := uint64(time.Now().UnixNano())
	if nw > ak.Expiration {
		return errors.New("Node account key has expired")
	}
	a := db.GetAccount(ak.AccountID)
	if a == nil {
		return errors.New("Account could not be found")
	}
	if !a.Enabled {
		return errors.New("Account is not enabled")
	}
	if !a.AllowNodeAcocuntKey {
		return errors.New("Account does not allow node account keys")
	}
	if a.NumberNodes >= a.MaxNodes {
		return errors.New("Maximum number of nodes for account reached")
	}
	err = db.IncrNumberNodes(a)
	if err != nil {
		return err
	}
	var na gripdata.NodeAccount
	na.AccountID = a.AccountID
	na.Enabled = true
	na.NodeID = s.NodeID
	err = db.StoreNodeAccount(&na)
	if err != nil {
		return err
	}
	return nil
}
