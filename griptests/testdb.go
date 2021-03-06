package griptests

import (
	"encoding/base64"
	"log"
	"reflect"
	"sync"

	"github.com/wyathan/grip"
	"github.com/wyathan/grip/gripcrypto"
	"github.com/wyathan/grip/gripdata"
)

//TestDB is that
type TestDB struct {
	sync.Mutex
	LclNode              *gripdata.Node
	LclPrvNodeData       *gripdata.MyNodePrivateData
	Accounts             map[string]*gripdata.Account
	AccountKeys          map[string]*gripdata.NodeAccountKey
	NodeAccounts         map[string]*gripdata.NodeAccount
	Nodes                map[string]*gripdata.Node
	ShareNodes           map[string][]gripdata.ShareNodeInfo
	ShareNodeKeys        map[string][]gripdata.ShareNodeInfo
	UseShareKeys         map[string][]gripdata.UseShareNodeKey
	SendData             map[string][]gripdata.SendData
	DigData              map[string]interface{}
	NodeEphemera         map[string]*gripdata.NodeEphemera
	Contexts             map[string]*gripdata.Context
	ContextRequests      map[string][]*gripdata.ContextRequest
	ContextResponses     map[string]map[string]*gripdata.ContextResponse
	ContextFiles         map[string][]gripdata.ContextFileWrap
	ContextFilesByDepDig map[string]*gripdata.ContextFileWrap
	RejectedData         map[string][]gripdata.RejectedSendData
	FileTransfers        map[string][]*gripdata.ContextFileTransferWrap
	DeletedFiles         map[string]*gripdata.DeletedContextFile
	VeryBadContextFiles  []gripdata.ContextFile
}

func NewTestDB() *TestDB {
	var t TestDB
	t.DeletedFiles = make(map[string]*gripdata.DeletedContextFile)
	t.FileTransfers = make(map[string][]*gripdata.ContextFileTransferWrap)
	t.Accounts = make(map[string]*gripdata.Account)
	t.AccountKeys = make(map[string]*gripdata.NodeAccountKey)
	t.NodeAccounts = make(map[string]*gripdata.NodeAccount)
	t.Nodes = make(map[string]*gripdata.Node)
	t.ShareNodes = make(map[string][]gripdata.ShareNodeInfo)
	t.ShareNodeKeys = make(map[string][]gripdata.ShareNodeInfo)
	t.UseShareKeys = make(map[string][]gripdata.UseShareNodeKey)
	t.SendData = make(map[string][]gripdata.SendData)
	t.DigData = make(map[string]interface{})
	t.NodeEphemera = make(map[string]*gripdata.NodeEphemera)
	t.Contexts = make(map[string]*gripdata.Context)
	t.ContextRequests = make(map[string][]*gripdata.ContextRequest)
	t.ContextResponses = make(map[string]map[string]*gripdata.ContextResponse)
	t.ContextFiles = make(map[string][]gripdata.ContextFileWrap)
	t.RejectedData = make(map[string][]gripdata.RejectedSendData)
	t.ContextFilesByDepDig = make(map[string]*gripdata.ContextFileWrap)
	return &t
}

func (t *TestDB) addDig(s gripcrypto.SignInf) {
	if t.DigData == nil {
		t.DigData = make(map[string]interface{})
	}
	ds := base64.StdEncoding.EncodeToString(s.GetDig())
	log.Printf("ADDDIG %s %s\n", ds, reflect.TypeOf(s).String())
	t.DigData[ds] = s
}

//testTestDBImplements make sure we implement the interfaces
func testTestDBImplements() {
	var t TestDB
	var s grip.DB
	s = &t
	s.GetAccount("")
}
