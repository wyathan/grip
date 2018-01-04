package griptests

import (
	"crypto/rand"
	"encoding/base64"
	"io/ioutil"
	"log"
	"testing"

	"github.com/wyathan/grip"
	"github.com/wyathan/grip/gripdata"
)

func countContextRequests(m map[string][]*gripdata.ContextRequest) int {
	cnt := 0
	for _, rl := range m {
		cnt += len(rl)
	}
	return cnt
}

func countContextResponses(m map[string]map[string]*gripdata.ContextResponse) int {
	cnt := 0
	for _, rl := range m {
		cnt += len(rl)
	}
	return cnt
}

func countContextFiles(m map[string][]gripdata.ContextFileWrap) int {
	cnt := 0
	for _, rl := range m {
		cnt += len(rl)
	}
	return cnt
}

func MakeTempFile() string {
	f, _ := ioutil.TempFile("", "grip")
	defer f.Close()
	randdata := make([]byte, 1024)
	rand.Read(randdata)
	f.Write(randdata)
	return f.Name()
}

func TestContexts(t *testing.T) {
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

	for cnt := 2; cnt < 5; cnt++ {
		var ks gripdata.UseShareNodeKey
		ks.Key = shr.Key
		ks.TargetID = nodes[0].ID
		err = grip.NewUseShareNodeKey(&ks, dbs[cnt])
		if err != nil {
			t.Error(err)
		}
	}
	if !WaitUntilAllSent(dbs) {
		t.Error("Failed to send all")
	}

	for cnt := 1; cnt < 5; cnt++ {
		if 5 != len(dbs[cnt].Nodes) {
			t.Errorf("Node %d only knows %d other nodes", cnt, len(dbs[cnt].Nodes))
		}
	}

	var ctx gripdata.Context
	ctx.Name = "testcontext00"
	err = grip.NewContext(&ctx, dbs[1])
	if err != nil {
		t.Error("Failed to create new context")
	}

	for cnt := 0; cnt < 5; cnt++ {
		if cnt == 1 {
			continue
		}
		var rq gripdata.ContextRequest
		rq.ContextDig = ctx.Dig
		rq.TargetNodeID = nodes[cnt].ID
		err = grip.NewContextRequest(&rq, dbs[1])
		if err != nil {
			t.Errorf("Failed to create context request: %s", err)
		}
	}

	if !WaitUntilAllSent(dbs) {
		t.Error("Failed to send all")
	}

	for cnt := 0; cnt < 5; cnt++ {
		if 4 != countContextRequests(dbs[cnt].ContextRequests) {
			t.Errorf("Node %d only has %d context requests", cnt, countContextRequests(dbs[cnt].ContextRequests))
		}
		if 1 != countContextResponses(dbs[cnt].ContextResponses) {
			t.Errorf("Node %d only has %d context responses", cnt, countContextResponses(dbs[cnt].ContextResponses))
		}
		if 1 != len(dbs[cnt].Contexts) {
			t.Errorf("Node %d does not know the context", cnt)
		}
	}

	for cnt := 5; cnt < 10; cnt++ {
		if 0 != countContextRequests(dbs[cnt].ContextRequests) {
			t.Errorf("Node %d knows %d context requests", cnt, countContextRequests(dbs[cnt].ContextRequests))
		}
		if 0 != len(dbs[cnt].Contexts) {
			t.Errorf("Node %d knows the context", cnt)
		}
		if 0 != countContextResponses(dbs[cnt].ContextResponses) {
			t.Errorf("Node %d has %d context responses", cnt, countContextResponses(dbs[cnt].ContextResponses))
		}
	}

	var rsp gripdata.ContextResponse
	rsp.ContextDig = ctx.Dig
	err = grip.NewContextResponse(&rsp, dbs[2])
	if err != nil {
		t.Errorf("Failed to create context response %s", err)
	}
	if !WaitUntilAllSent(dbs) {
		t.Error("Failed to send all")
	}

	for cnt := 0; cnt < 5; cnt++ {
		if 4 != countContextRequests(dbs[cnt].ContextRequests) {
			t.Errorf("Node %d only has %d context requests", cnt, countContextRequests(dbs[cnt].ContextRequests))
		}
		if 2 != countContextResponses(dbs[cnt].ContextResponses) {
			t.Errorf("Node %d only has %d context responses", cnt, countContextResponses(dbs[cnt].ContextResponses))
		}
		if 1 != len(dbs[cnt].Contexts) {
			t.Errorf("Node %d does not know the context", cnt)
		}
	}

	for cnt := 5; cnt < 10; cnt++ {
		if 0 != countContextRequests(dbs[cnt].ContextRequests) {
			t.Errorf("Node %d knows %d context requests", cnt, countContextRequests(dbs[cnt].ContextRequests))
		}
		if 0 != len(dbs[cnt].Contexts) {
			t.Errorf("Node %d knows the context", cnt)
		}
		if 0 != countContextResponses(dbs[cnt].ContextResponses) {
			t.Errorf("Node %d has %d context responses", cnt, countContextResponses(dbs[cnt].ContextResponses))
		}
	}

	var f gripdata.ContextFile
	f.Context = ctx.Dig
	f.Path = MakeTempFile()
	err = grip.NewContextFile(&f, dbs[1])
	if err != nil {
		t.Errorf("Failed to create new context file: %s", err)
	}

	if !WaitUntilAllSent(dbs) {
		t.Error("Failed to send all")
	}

	for cnt := 0; cnt < 3; cnt++ {
		if 1 != countContextFiles(dbs[cnt].ContextFiles) {
			t.Errorf("Node %d knows %d context files", cnt, countContextFiles(dbs[cnt].ContextFiles))
		}
	}
	for cnt := 3; cnt < 10; cnt++ {
		if 0 != countContextFiles(dbs[cnt].ContextFiles) {
			t.Errorf("Node %d knows %d context files", cnt, countContextFiles(dbs[cnt].ContextFiles))
		}
	}

	tn.CloseAll()
}
