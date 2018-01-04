package griptests

import (
	"bytes"
	"encoding/base64"

	"github.com/wyathan/grip/gripdata"
)

/*
	StoreContext(c *gripdata.Context) error
	GetContext(id []byte) *gripdata.Context
	GetContextRequests(id []byte) []gripdata.ContextRequest
	StoreContextRequest(c *gripdata.ContextRequest) error
	GetContextRequest(cid []byte, tgtid []byte) *gripdata.ContextRequest
	StoreContextResponse(c *gripdata.ContextResponse) error
	GetContextResponse(cid []byte, tgtid []byte) *gripdata.ContextResponse
	GetContextResponses(cid []byte) []gripdata.ContextResponse
	GetContextFileByDepDataDig(d []byte) *gripdata.ContextFile
	StoreVeryBadContextFile(c *gripdata.ContextFile) error
	StoreContextFile(c *gripdata.ContextFile) error
	StoreContextFileTransfer(c *gripdata.ContextFileTransfer) error
	SetContextHeadFile(c *gripdata.ContextFile) error
	RemoveContextHeadFile(dig []byte) error
	GetAllContextHeadFiles(cid []byte, index bool) []gripdata.ContextFile
*/

func (t *TestDB) StoreContext(c *gripdata.Context) error {
	t.Lock()
	defer t.Unlock()
	id := base64.StdEncoding.EncodeToString(c.Dig)
	t.Contexts[id] = c
	t.addDig(c)
	return nil
}
func (t *TestDB) GetContext(id []byte) *gripdata.Context {
	t.Lock()
	defer t.Unlock()
	sid := base64.StdEncoding.EncodeToString(id)
	return t.Contexts[sid]
}
func (t *TestDB) GetContextRequests(id []byte) []*gripdata.ContextRequest {
	t.Lock()
	defer t.Unlock()
	sid := base64.StdEncoding.EncodeToString(id)
	return t.ContextRequests[sid]
}
func (t *TestDB) StoreContextRequest(c *gripdata.ContextRequest) error {
	t.Lock()
	defer t.Unlock()
	sid := base64.StdEncoding.EncodeToString(c.ContextDig)
	t.ContextRequests[sid] = append(t.ContextRequests[sid], c)
	t.addDig(c)
	return nil
}
func (t *TestDB) GetContextRequest(cid []byte, tgtid []byte) *gripdata.ContextRequest {
	t.Lock()
	defer t.Unlock()
	scid := base64.StdEncoding.EncodeToString(cid)
	tar := t.ContextRequests[scid]
	for _, cr := range tar {
		if bytes.Equal(tgtid, cr.TargetNodeID) {
			return cr
		}
	}
	return nil
}
func (t *TestDB) StoreContextResponse(c *gripdata.ContextResponse) error {
	t.Lock()
	defer t.Unlock()
	cid := base64.StdEncoding.EncodeToString(c.ContextDig)
	tid := base64.StdEncoding.EncodeToString(c.TargetNodeID)
	m := t.ContextResponses[cid]
	if m == nil {
		m = make(map[string]*gripdata.ContextResponse)
		t.ContextResponses[cid] = m
	}
	m[tid] = c
	t.addDig(c)
	return nil
}
func (t *TestDB) GetContextResponse(cid []byte, tgtid []byte) *gripdata.ContextResponse {
	t.Lock()
	defer t.Unlock()
	scid := base64.StdEncoding.EncodeToString(cid)
	stid := base64.StdEncoding.EncodeToString(tgtid)
	rsp := t.ContextResponses[scid]
	if rsp != nil {
		return rsp[stid]
	}
	return nil
}
func (t *TestDB) GetContextFileByDepDataDig(d []byte) *gripdata.ContextFileWrap {
	t.Lock()
	defer t.Unlock()
	dd := base64.StdEncoding.EncodeToString(d)
	return t.ContextFilesByDepDig[dd]
}
func (t *TestDB) GetContextResponses(cid []byte) []*gripdata.ContextResponse {
	t.Lock()
	defer t.Unlock()
	sid := base64.StdEncoding.EncodeToString(cid)
	rm := t.ContextResponses[sid]
	var r []gripdata.ContextResponse
	for _, v := range rm {
		r = append(r, *v)
	}
	return nil
}
func (t *TestDB) StoreContextFile(cf *gripdata.ContextFile) error {
	t.Lock()
	defer t.Unlock()
	cid := base64.StdEncoding.EncodeToString(cf.Context)
	dd := base64.StdEncoding.EncodeToString(cf.DataDepDig)
	var c gripdata.ContextFileWrap
	c.ContextFile = cf
	t.ContextFilesByDepDig[dd] = &c
	fl := t.ContextFiles[cid]
	t.ContextFiles[cid] = append(fl, c)
	t.addDig(c.ContextFile)
	return c.UpdateContextFileWrapDB(t)
}
func (t *TestDB) StoreVeryBadContextFile(c *gripdata.ContextFile) error {
	t.Lock()
	defer t.Unlock()
	t.VeryBadContextFiles = append(t.VeryBadContextFiles, *c)
	return nil
}
func (t *TestDB) GetAllThatDependOn(cid []byte, dig []byte) []*gripdata.ContextFileWrap {
	t.Lock()
	defer t.Unlock()
	return t.TGetAllThatDependOn(cid, dig)
}
func (t *TestDB) StoreContextFileTransfer(c *gripdata.ContextFileTransfer) error {
	t.Lock()
	defer t.Unlock()
	scid := base64.StdEncoding.EncodeToString(c.TrasnferTo)
	fl := t.FileTransfers[scid]
	fl = append(fl, *c)
	t.FileTransfers[scid] = fl
	t.addDig(c)
	return nil
}
func (t *TestDB) DeleteContextFile(c *gripdata.ContextFileWrap) error {
	t.Lock()
	defer t.Unlock()
	scid := base64.StdEncoding.EncodeToString(c.ContextFile.Context)
	fl := t.ContextFiles[scid]
	var nl []gripdata.ContextFileWrap
	for _, f := range fl {
		if !bytes.Equal(c.ContextFile.Dig, f.ContextFile.Dig) {
			nl = append(nl, f)
		}
	}
	t.ContextFiles[scid] = nl
	scid = base64.StdEncoding.EncodeToString(c.ContextFile.DataDepDig)
	delete(t.ContextFilesByDepDig, scid)
	return nil
}
func (t *TestDB) TGetAllThatDependOn(cid []byte, dig []byte) []*gripdata.ContextFileWrap {
	scid := base64.StdEncoding.EncodeToString(cid)
	var r []*gripdata.ContextFileWrap
	for _, fl := range t.ContextFiles[scid] {
		if gripdata.DoesDependOn(dig, fl.ContextFile) {
			r = append(r, &fl)
		}
	}
	return r
}
func (t *TestDB) TGetContextFileByDepDataDig(dd []byte) *gripdata.ContextFileWrap {
	return nil
}
func (t *TestDB) TUpdateContextFileWrap(c *gripdata.ContextFileWrap) error {
	return nil
}

//testTestDBImplements make sure we implement the interfaces
func testTestDBImplementsContextFileWrapdb() {
	var t TestDB
	var s gripdata.ContextFileWrapdb
	s = &t
	s.TGetAllThatDependOn(nil, nil)
}
