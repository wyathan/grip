package griptests

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/wyathan/grip"
	"github.com/wyathan/grip/gripdata"
)

func createNewNode(idx int, c bool, tn *TestNetwork) (*gripdata.MyNodePrivateData, *gripdata.Node, *TestDB) {
	var n gripdata.Node
	var pn gripdata.MyNodePrivateData
	n.Connectable = c
	tdb := NewTestDB()
	grip.CreateNewNode(&pn, &n, tdb)
	sk := tn.Open(n.ID, idx)
	sctrl := grip.NewSocketController(sk, tdb)
	sctrl.Start()
	return &pn, &n, tdb
}

//TestNodeShare does that
func TestNodeShare(t *testing.T) {
	log.Printf("HERE")
	var nodes []*gripdata.Node
	var pnodes []*gripdata.MyNodePrivateData
	var dbs []*TestDB
	tn := InitTestNetwork()
	for c := 0; c < 10; c++ {
		pr, n, db := createNewNode(c, c == 0, tn)
		nodes = append(nodes, n)
		dbs = append(dbs, db)
		pnodes = append(pnodes, pr)
	}
	pnodes[0].AutoShareNodeInfo = false

	for c := 1; c < 10; c++ {
		grip.IncomingNode(nodes[0], dbs[c])
		var a gripdata.Account
		var na gripdata.NodeAccount
		a.AccountID = fmt.Sprintf("node%d", c)
		a.Enabled = true
		a.AllowNodeAcocuntKey = true
		dbs[0].StoreAccount(&a)

		na.AccountID = a.AccountID
		na.Enabled = true
		na.NodeID = nodes[c].ID
		dbs[0].StoreNodeAccount(&na)

		pnodes[c].AutoCreateShareAccount = true
		pnodes[c].AutoAccountAllowContextNode = true
		pnodes[c].AutoAccountAllowContextSource = true
		pnodes[c].AutoAccountAllowFullRepo = true
		pnodes[c].AutoAccountAllowNodeAcocuntKey = true
		pnodes[c].AutoAccountMaxNodes = 1000
		pnodes[c].AutoAccountMaxDiskSpace = 1024 * 1024 * 1024
		dbs[c].StoreMyPrivateNodeData(nodes[c], pnodes[c])
	}

	var shr gripdata.ShareNodeInfo
	shr.Key = "abcd123"
	shr.NodeID = nodes[1].ID
	shr.TargetNodeID = nodes[0].ID
	err := grip.NewShareNode(&shr, dbs[1])
	if err != nil {
		t.Error(err)
	}
	time.Sleep(2 * time.Second)
	var ks gripdata.UseShareNodeKey
	ks.Key = shr.Key
	ks.TargetID = nodes[0].ID
	err = grip.NewUseShareNodeKey(&ks, dbs[2])
	if err != nil {
		t.Error(err)
	}
	time.Sleep(5 * time.Second)

	if 3 != len(dbs[0].Nodes) {
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

	tn.CloseAll()
}
