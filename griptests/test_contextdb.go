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
func (t *TestDB) GetContextRequests(id []byte) []gripdata.ContextRequest {
	t.Lock()
	defer t.Unlock()
	sid := base64.StdEncoding.EncodeToString(id)
	return t.ContextRequests[sid]
}
func (t *TestDB) StoreContextRequest(c *gripdata.ContextRequest) error {
	t.Lock()
	defer t.Unlock()
	sid := base64.StdEncoding.EncodeToString(c.ContextDig)
	t.ContextRequests[sid] = append(t.ContextRequests[sid], *c)
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
			return &cr
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
func (t *TestDB) GetContextFileByDepDataDig(d []byte) *gripdata.ContextFile {
	t.Lock()
	defer t.Unlock()
	dd := base64.StdEncoding.EncodeToString(d)
	return t.ContextFilesByDepDig[dd]
}
func (t *TestDB) GetContextResponses(cid []byte) []gripdata.ContextResponse {
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
func (t *TestDB) StoreContextFile(c *gripdata.ContextFile) error {
	t.Lock()
	defer t.Unlock()
	cid := base64.StdEncoding.EncodeToString(c.Context)
	dd := base64.StdEncoding.EncodeToString(c.DataDepDig)
	t.ContextFilesByDepDig[dd] = c
	fl := t.ContextFiles[cid]
	t.ContextFiles[cid] = append(fl, *c)
	t.addDig(c)
	return nil
}
func (t *TestDB) StoreVeryBadContextFile(c *gripdata.ContextFile) error {
	t.Lock()
	defer t.Unlock()
	t.VeryBadContextFiles = append(t.VeryBadContextFiles, *c)
	return nil
}
func (t *TestDB) GetAllThatDependOn(cid []byte, dig []byte) []gripdata.ContextFile {
	var r []gripdata.ContextFile
	return r
}
