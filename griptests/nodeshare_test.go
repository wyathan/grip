package griptests

import (
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"testing"
	"time"

	"github.com/wyathan/grip"
	"github.com/wyathan/grip/gripdata"
)

var NODEMAP = make(map[string]int)
var SOCKETS []*grip.SocketController

func clearTestGlobals() {
	NODEMAP = make(map[string]int)
	SOCKETS = nil
}

func createNewNode(idx int, c bool, tn *TestNetwork) (*gripdata.MyNodePrivateData, *gripdata.Node, *TestDB) {
	var n gripdata.Node
	var pn gripdata.MyNodePrivateData
	n.Connectable = c
	tdb := NewTestDB()
	grip.CreateNewNode(&pn, &n, tdb)
	NODEMAP[base64.StdEncoding.EncodeToString(n.ID)] = idx
	sk := tn.Open(n.ID, idx)
	sctrl := grip.NewSocketController(sk, tdb)
	sctrl.Start()
	SOCKETS = append(SOCKETS, sctrl)
	return &pn, &n, tdb
}

func NumToSendTo(db *TestDB, nid int) int {
	for k, v := range db.SendData {
		if NODEMAP[k] == nid {
			return len(v)
		}
	}
	return 0
}

func WaitUntilZero(db *TestDB, nid int) bool {
	lastsize := 0
	numtosend := NumToSendTo(db, nid)
	failloop := 0
	for maxloops := 30; numtosend > 0 && maxloops > 0; maxloops-- {
		time.Sleep(30 * time.Second)
		numtosend = NumToSendTo(db, nid)
		if lastsize > 0 && numtosend == lastsize {
			failloop++
			if failloop > 25 {
				log.Printf("has not sent any in the last 30 seconds")
				return false
			}
		} else {
			failloop = 0
		}
		lastsize = numtosend
	}
	return NumToSendTo(db, nid) == 0
}

func WaitUntilAllSent(dbs []*TestDB) bool {
	for i := 1; i < len(dbs); i++ {
		if !WaitUntilZero(dbs[0], i) {
			log.Printf("Zero failed to send all to %d", i)
			return false
		}
		if !WaitUntilZero(dbs[i], 0) {
			log.Printf("%d failed to send all to zero", i)
			return false
		}
	}
	return true
}

func createSomeNodes(num int) (tn *TestNetwork, nodes []*gripdata.Node, pnodes []*gripdata.MyNodePrivateData, dbs []*TestDB) {
	clearTestGlobals()
	tn = InitTestNetwork()
	ct := uint64(time.Now().UnixNano()) + (10 * uint64(time.Minute))
	for c := 0; c < num; c++ {
		pr, n, db := createNewNode(c, c == 0, tn)
		nodes = append(nodes, n)
		dbs = append(dbs, db)
		pnodes = append(pnodes, pr)
	}
	pnodes[0].AutoShareNodeInfo = false

	for c := 1; c < num; c++ {
		grip.IncomingNode(nodes[0], dbs[c])
		var a gripdata.Account
		var na gripdata.NodeAccountKey
		a.AccountID = fmt.Sprintf("node%d", c)
		a.Enabled = true
		a.AllowNodeAcocuntKey = true
		a.MaxNodes = 1
		dbs[0].StoreAccount(&a)

		na.AccountID = a.AccountID
		na.OneTime = true
		na.Expiration = ct
		na.Key = fmt.Sprintf("node%dkey", c)
		dbs[0].StoreNodeAccountKey(&na)

		pnodes[c].AutoCreateShareAccount = true
		pnodes[c].AutoAccountAllowContextNode = true
		pnodes[c].AutoAccountAllowContextSource = true
		pnodes[c].AutoAccountAllowFullRepo = true
		pnodes[c].AutoAccountAllowNodeAcocuntKey = true
		pnodes[c].AutoAccountMaxNodes = 1000
		pnodes[c].AutoAccountMaxDiskSpace = 1024 * 1024 * 1024
		dbs[c].StoreMyPrivateNodeData(nodes[c], pnodes[c])

		grip.AssociateNodeAccoutKey(na.Key, nodes[0].ID, dbs[c])
	}
	if !WaitUntilAllSent(dbs) {
		log.Fatal("Failed to send assocate node key")
	}
	log.Printf("<<<<<<<<<<<<<<<<<<<<<<< %d nodes created <<<<<<<<<<<<<<<<<<<<<<<<<", num)
	return tn, nodes, pnodes, dbs
}

//TestNodeShare does that
func TestNodeShare(t *testing.T) {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	seedv := time.Now().UnixNano()
	log.Printf("SEED VALUE: %d", seedv)
	rand.Seed(seedv)

	tn, nodes, _, dbs := createSomeNodes(10)

	var shr gripdata.ShareNodeInfo
	shr.Key = "abcd123"
	shr.NodeID = nodes[1].ID
	shr.TargetNodeID = nodes[0].ID
	err := grip.NewShareNode(&shr, dbs[1])
	if err != nil {
		t.Error(err)
	}
	log.Printf("Created node 1 to node 0 ShareNodeInfo: %s\n", base64.StdEncoding.EncodeToString(shr.GetDig()))
	if !WaitUntilAllSent(dbs) {
		t.Error("Failed to send all")
	}

	var ks gripdata.UseShareNodeKey
	ks.Key = shr.Key
	ks.TargetID = nodes[0].ID
	err = grip.NewUseShareNodeKey(&ks, dbs[2])
	if err != nil {
		t.Error(err)
	}
	log.Printf("Created node 2 to node 0 UseShareNodeKey: %s\n", base64.StdEncoding.EncodeToString(ks.GetDig()))
	if !WaitUntilAllSent(dbs) {
		t.Error("Failed to send all")
	}

	if 1 != len(dbs[0].UseShareKeys[shr.Key]) {
		t.Error("Node 0 did not get the UseShareNodeKey from node 2")
	}

	for t, sdl := range dbs[0].SendData {
		log.Printf("Node 0 has Data to send to %d", NODEMAP[t])
		for _, sd := range sdl {
			log.Printf("      %s", base64.StdEncoding.EncodeToString(sd.Dig))
		}
	}
	if 10 != len(dbs[0].Nodes) {
		for _, kn := range dbs[0].Nodes {
			log.Printf("Node 0 knows node: %d\n", NODEMAP[base64.StdEncoding.EncodeToString(kn.ID)])
		}
		t.Errorf("Missing nodes %d", len(dbs[0].Nodes))
	}
	if 3 != len(dbs[1].Nodes) {
		for _, kn := range dbs[1].Nodes {
			log.Printf("Node 1 knows node: %d\n", NODEMAP[base64.StdEncoding.EncodeToString(kn.ID)])
		}
		t.Errorf("Missing nodes %d", len(dbs[1].Nodes))
	}
	if 3 != len(dbs[2].Nodes) {
		for _, kn := range dbs[2].Nodes {
			log.Printf("Node 2 knows node: %d\n", NODEMAP[base64.StdEncoding.EncodeToString(kn.ID)])
		}
		t.Errorf("Missing nodes %d", len(dbs[2].Nodes))
	}
	if 2 != len(dbs[3].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[3].Nodes))
	}
	if 2 != len(dbs[4].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[4].Nodes))
	}
	if 2 != len(dbs[5].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[5].Nodes))
	}
	if 2 != len(dbs[6].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[6].Nodes))
	}
	if 2 != len(dbs[7].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[7].Nodes))
	}
	if 2 != len(dbs[8].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[8].Nodes))
	}
	if 2 != len(dbs[9].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[9].Nodes))
	}

	var shr3 gripdata.ShareNodeInfo
	shr3.Key = "node3sharekey"
	shr3.NodeID = nodes[3].ID
	shr3.TargetNodeID = nodes[0].ID
	err = grip.NewShareNode(&shr3, dbs[3])
	if err != nil {
		t.Error(err)
	}
	log.Printf("Created node 3 to node 0 ShareNodeInfo: %s\n", base64.StdEncoding.EncodeToString(shr3.GetDig()))

	if !WaitUntilAllSent(dbs) {
		t.Error("Failed to send all")
	}

	if 10 != len(dbs[0].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[0].Nodes))
	}
	if 3 != len(dbs[1].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[1].Nodes))
	}
	if 3 != len(dbs[2].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[2].Nodes))
	}
	if 2 != len(dbs[3].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[3].Nodes))
	}
	if 2 != len(dbs[4].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[4].Nodes))
	}
	if 2 != len(dbs[5].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[5].Nodes))
	}
	if 2 != len(dbs[6].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[6].Nodes))
	}
	if 2 != len(dbs[7].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[7].Nodes))
	}
	if 2 != len(dbs[8].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[8].Nodes))
	}
	if 2 != len(dbs[9].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[9].Nodes))
	}

	//var ks2 gripdata.UseShareNodeKey
	//ks2.Key = shr.Key
	//ks2.TargetID = nodes[0].ID
	//err = grip.NewUseShareNodeKey(&ks2, dbs[3])
	var ks2 gripdata.UseShareNodeKey
	ks2.Key = "node3sharekey"
	ks2.TargetID = nodes[0].ID
	err = grip.NewUseShareNodeKey(&ks2, dbs[1])
	if err != nil {
		t.Error(err)
	}
	log.Printf("Created node 1 to node 3 UseShareNodeKey: %s\n", base64.StdEncoding.EncodeToString(ks2.GetDig()))

	if !WaitUntilAllSent(dbs) {
		t.Error("Failed to send all")
	}

	if 10 != len(dbs[0].Nodes) { //0,1,2,3
		t.Errorf("Missing nodes %d", len(dbs[0].Nodes))
	}
	if 4 != len(dbs[1].Nodes) { //0,1,2,3
		for _, kn := range dbs[1].Nodes {
			log.Printf("Node 1 knows node: %d\n", NODEMAP[base64.StdEncoding.EncodeToString(kn.ID)])
		}
		t.Errorf("Missing nodes %d", len(dbs[1].Nodes))
	}
	if 3 != len(dbs[2].Nodes) { //0,1,2
		t.Errorf("Missing nodes %d", len(dbs[2].Nodes))
	}
	if 3 != len(dbs[3].Nodes) { //0,1,3
		t.Errorf("Missing nodes %d", len(dbs[3].Nodes))
	}
	if 2 != len(dbs[4].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[4].Nodes))
	}
	if 2 != len(dbs[5].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[5].Nodes))
	}
	if 2 != len(dbs[6].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[6].Nodes))
	}
	if 2 != len(dbs[7].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[7].Nodes))
	}
	if 2 != len(dbs[8].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[8].Nodes))
	}
	if 2 != len(dbs[9].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[9].Nodes))
	}

	var shr4 gripdata.ShareNodeInfo
	shr4.Key = "another1key"
	shr4.NodeID = nodes[1].ID
	shr4.TargetNodeID = nodes[3].ID
	err = grip.NewShareNode(&shr4, dbs[1])
	if err != nil {
		t.Error(err)
	}
	log.Printf("Created node 1 to node 3 ShareNodeInfo: %s\n", base64.StdEncoding.EncodeToString(shr4.GetDig()))

	if !WaitUntilAllSent(dbs) {
		t.Error("Failed to send all")
	}

	for t, sdl := range dbs[0].SendData {
		log.Printf("Node 0 has Data to send to %d", NODEMAP[t])
		for _, sd := range sdl {
			log.Printf("      %s", base64.StdEncoding.EncodeToString(sd.Dig))
		}
	}
	if 10 != len(dbs[0].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[0].Nodes))
	}
	if 4 != len(dbs[1].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[1].Nodes))
	}
	if 4 != len(dbs[2].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[2].Nodes))
	}
	if 4 != len(dbs[3].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[3].Nodes))
	}
	if 2 != len(dbs[4].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[4].Nodes))
	}
	if 2 != len(dbs[5].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[5].Nodes))
	}
	if 2 != len(dbs[6].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[6].Nodes))
	}
	if 2 != len(dbs[7].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[7].Nodes))
	}
	if 2 != len(dbs[8].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[8].Nodes))
	}
	if 2 != len(dbs[9].Nodes) {
		t.Errorf("Missing nodes %d", len(dbs[9].Nodes))
	}

	tn.CloseAll()
}
