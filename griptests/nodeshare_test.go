package griptests

import (
	"log"
	"testing"
	"time"

	"github.com/wyathan/grip"
	"github.com/wyathan/grip/gripdata"
)

func createNewNode(idx int, tn *TestNetwork) (*gripdata.MyNodePrivateData, *gripdata.Node, grip.NodeNetAccountdb) {
	var n gripdata.Node
	var pn gripdata.MyNodePrivateData
	n.Connectable = true
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
	var dbs []grip.NodeNetAccountdb
	tn := InitTestNetwork()
	for c := 0; c < 10; c++ {
		pr, n, db := createNewNode(c, tn)
		nodes = append(nodes, n)
		dbs = append(dbs, db)
		pnodes = append(pnodes, pr)
	}
	pnodes[0].AutoShareNodeInfo = true

	for c := 1; c < 10; c++ {
		grip.IncomingNode(nodes[0], dbs[c])
	}

	var shr gripdata.ShareNodeInfo
	shr.Key = "abcd123"
	shr.NodeID = nodes[1].ID
	shr.TargetNodeID = nodes[0].ID
	err := grip.NewShareNode(&shr, dbs[1])
	if err != nil {
		t.Error(err)
	}
	time.Sleep(5 * time.Second)
	tn.CloseAll()
}
