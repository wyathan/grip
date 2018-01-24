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
	StoreContextFile(c *gripdata.ContextFile, s int64) (*gripdata.ContextFileWrap, error)
	StoreContextFileTransfer(c *gripdata.ContextFileTransfer) (*gripdata.ContextFileTransferWrap, error)
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
	return t.TGetContextFileByDepDataDig(d)
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
func (t *TestDB) StoreContextFile(cf *gripdata.ContextFile) (*gripdata.ContextFileWrap, error) {
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
	return &c, c.UpdateContextFileWrapDB(t)
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
func (t *TestDB) StoreContextFileTransfer(c *gripdata.ContextFileTransfer) (*gripdata.ContextFileTransferWrap, error) {
	t.Lock()
	defer t.Unlock()
	var w gripdata.ContextFileTransferWrap
	w.ContextFileTransfer = c
	scid := base64.StdEncoding.EncodeToString(c.TrasnferTo)
	fl := t.FileTransfers[scid]
	fl = append(fl, &w)
	t.FileTransfers[scid] = fl
	t.addDig(c)
	return &w, nil
}
func (t *TestDB) DeleteContextFileTransfer(nodeid []byte, confiledig []byte) (string, error) {
	t.Lock()
	defer t.Unlock()
	snid := base64.StdEncoding.EncodeToString(nodeid)
	var tf []*gripdata.ContextFileTransferWrap
	fl := t.FileTransfers[snid]
	if fl != nil {
		for _, w := range fl {
			if !bytes.Equal(w.ContextFileTransfer.ContextFileDig, confiledig) {
				tf = append(tf, w)
			}
		}
	}
	t.FileTransfers[snid] = tf
	//Doesn't hurt to just say you can never delete for now.
	//TODO: Add function to check if we can delete the file
	return "", nil
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
	sdig := base64.StdEncoding.EncodeToString(c.ContextFile.Dig)
	var d gripdata.DeletedContextFile
	d.Context = c.ContextFile.Context
	d.DataDepDig = c.ContextFile.DataDepDig
	d.Dig = c.ContextFile.Dig
	t.DeletedFiles[sdig] = &d
	return nil
}
func (t *TestDB) TGetAllThatDependOn(cid []byte, dig []byte) []*gripdata.ContextFileWrap {
	scid := base64.StdEncoding.EncodeToString(cid)
	var r []*gripdata.ContextFileWrap
	for _, fl := range t.ContextFiles[scid] {
		if gripdata.DoesDependOn(dig, fl.ContextFile) {
			v := fl
			r = append(r, &v)
		}
	}
	return r
}
func (t *TestDB) TGetContextFileByDepDataDig(dd []byte) *gripdata.ContextFileWrap {
	sdd := base64.StdEncoding.EncodeToString(dd)
	return t.ContextFilesByDepDig[sdd]
}
func (t *TestDB) TUpdateContextFileWrap(cf *gripdata.ContextFileWrap) error {
	cid := base64.StdEncoding.EncodeToString(cf.ContextFile.Context)
	dd := base64.StdEncoding.EncodeToString(cf.ContextFile.DataDepDig)
	t.ContextFilesByDepDig[dd] = cf
	fl := t.ContextFiles[cid]
	fnd := false
	for c := 0; c < len(fl); c++ {
		if bytes.Equal(cf.ContextFile.Dig, fl[c].ContextFile.Dig) {
			fl[c] = *cf
			fnd = true
		}
	}
	if !fnd {
		t.ContextFiles[cid] = append(fl, *cf)
	}
	t.addDig(cf.ContextFile)
	return nil
}
func (t *TestDB) GetContextHeads(cid []byte) []*gripdata.ContextFileWrap {
	var r []*gripdata.ContextFileWrap
	sc := base64.StdEncoding.EncodeToString(cid)
	fl := t.ContextFiles[sc]
	for _, d := range fl {
		if d.Head {
			v := d
			r = append(r, &v)
		}
	}
	return r
}
func (t *TestDB) GetContextLeaves(cid []byte, covered bool, index bool) []*gripdata.ContextFileWrap {
	var r []*gripdata.ContextFileWrap
	sc := base64.StdEncoding.EncodeToString(cid)
	fl := t.ContextFiles[sc]
	for _, d := range fl {
		if d.Leaf && d.CoveredBySnapshot == covered && d.ContextFile.Index == index {
			v := d
			r = append(r, &v)
		}
	}
	r = mergeSortSize(r)
	r = mergeSortDepth(r)
	return r
}
func (t *TestDB) GetCoveredSnapshots(cid []byte) []*gripdata.ContextFileWrap {
	var r []*gripdata.ContextFileWrap
	sc := base64.StdEncoding.EncodeToString(cid)
	fl := t.ContextFiles[sc]
	for _, d := range fl {
		if d.ContextFile.Snapshot && d.CoveredBySnapshot {
			v := d
			r = append(r, &v)
		}
	}
	r = mergeSortSize(r)
	r = mergeSortDepth(r)
	return r
}
func (t *TestDB) GetFileTransfersForNode(id []byte, max int) []*gripdata.ContextFileTransferWrap {
	scid := base64.StdEncoding.EncodeToString(id)
	fl := t.FileTransfers[scid]
	if len(fl) < max {
		max = len(fl)
	}
	return fl[0:max]
}
func (t *TestDB) GetContextFileDeleted(dig []byte) *gripdata.DeletedContextFile {
	scid := base64.StdEncoding.EncodeToString(dig)
	return t.DeletedFiles[scid]
}

type depthCompare struct{}

func (d *depthCompare) Compare(a *gripdata.ContextFileWrap, b *gripdata.ContextFileWrap) int {
	return a.Depth - b.Depth
}

func mergeSortDepth(s []*gripdata.ContextFileWrap) []*gripdata.ContextFileWrap {
	return mergeSortContextFileWrap(s, &depthCompare{})
}

type sizeCompare struct{}

func (d *sizeCompare) Compare(a *gripdata.ContextFileWrap, b *gripdata.ContextFileWrap) int {
	return int(a.ContextFile.Size - b.ContextFile.Size)
}

func mergeSortSize(s []*gripdata.ContextFileWrap) []*gripdata.ContextFileWrap {
	return mergeSortContextFileWrap(s, &sizeCompare{})
}

type compareContextFileWrap interface {
	Compare(a *gripdata.ContextFileWrap, b *gripdata.ContextFileWrap) int
}

//Using merge sort because it's stable so we can sort size the depth
func mergeSortContextFileWrap(s []*gripdata.ContextFileWrap, cmp compareContextFileWrap) []*gripdata.ContextFileWrap {
	if len(s) <= 0 {
		return s
	}
	var mr [][]*gripdata.ContextFileWrap
	for c := 0; c < len(s); c++ {
		mr = append(mr, s[c:c+1])
	}
	for len(mr) > 1 {
		mr = append(doMergeSortContextFileWrap(mr[0], mr[1], cmp), mr[2:]...)

	}
	return mr[0]
}

func doMergeSortContextFileWrap(s0 []*gripdata.ContextFileWrap, s1 []*gripdata.ContextFileWrap, cmp compareContextFileWrap) [][]*gripdata.ContextFileWrap {
	var r []*gripdata.ContextFileWrap
	for len(s0) > 0 && len(s1) > 0 {
		if cmp.Compare(s0[0], s1[0]) >= 0 {
			r = append(r, s0[0])
			s0 = s0[1:]
		} else {
			r = append(r, s1[0])
			s1 = s1[1:]
		}
	}
	if len(s0) > 0 {
		r = append(r, s0...)
	}
	if len(s1) > 0 {
		r = append(r, s1...)
	}
	return [][]*gripdata.ContextFileWrap{r}
}

//testTestDBImplements make sure we implement the interfaces
func testTestDBImplementsContextFileWrapdb() {
	var t TestDB
	var s gripdata.ContextFileWrapdb
	s = &t
	s.TGetAllThatDependOn(nil, nil)
}
