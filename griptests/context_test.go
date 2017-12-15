package griptests

import (
	"testing"

	"github.com/wyathan/grip/gripdata"
)

func TestContexts(t *testing.T) {
	clearTestGlobals()

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

}
