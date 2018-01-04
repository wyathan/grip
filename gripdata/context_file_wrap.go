package gripdata

import (
	"bytes"
	"encoding/base64"
	"errors"
)

//ContextFileWrap wraps ContextFile to keep track
//of it's metadata within the local node.  It's specific
//to each node and never shared.
//The implementation of the Contextdb gets the honor of
//keeping track of this within transactions to update
//ContextFiles
type ContextFileWrap struct {
	ContextFile       *ContextFile
	Head              bool //is this file at the head of a dep tree
	Leaf              bool //is this file at the leaf/tail of dep tree path
	Depth             int  //how long is it from the head to this node
	CoveredBySnapshot bool //do all the dependencies of this file covered by snapshots
}

//ContextFileWrapdb database interface that can only be called
//while already within a transaction
type ContextFileWrapdb interface {
	TGetAllThatDependOn(cid []byte, dd []byte) []ContextFileWrap
	TGetContextFileByDepDataDig(dd []byte) *ContextFileWrap
}

//UpdateContextFileWrap MUST be executed within a transaction to update
//the context's file data.
func (c *ContextFileWrap) UpdateContextFileWrap(deps []*ContextFileWrap, dependon []*ContextFileWrap) error {
	err := sanityCheck(c.ContextFile, deps, dependon)
	if err != nil {
		return err
	}
	covered, depth := getDepthCovered(dependon)
	c.CoveredBySnapshot = covered
	c.Depth = depth
	c.Head = len(dependon) == 0
	c.Leaf = len(deps) == 0
	return nil
}

func getDepthCovered(dependon []*ContextFileWrap) (bool, int) {
	depth := 0
	covered := len(dependon) > 0
	for _, d := range dependon {
		if d.Depth+1 > depth {
			depth = d.Depth + 1
		}
		if !(d.CoveredBySnapshot || d.ContextFile.Snapshot) {
			covered = false
		}
	}
	return covered, depth
}

func sanityCheck(c *ContextFile, deps []*ContextFileWrap, dependon []*ContextFileWrap) error {
	err := sanityCheckDependson(c, dependon)
	if err != nil {
		return err
	}
	return sanityCheckDeps(c, deps)
}

func sanityCheckDependson(c *ContextFile, dependon []*ContextFileWrap) error {
	for _, d := range dependon {
		if !DoesDependOn(c.DataDepDig, d.ContextFile) {
			return errors.New("ContextFile in dependon slice does not depend on the ContextFile")
		}
	}
	return nil
}

//deps may not contain all the dependencies!  It is the list of
//dependencies we still have locally!
func sanityCheckDeps(c *ContextFile, deps []*ContextFileWrap) error {
	dmap := make(map[string]bool)
	for _, d := range c.DependsOn {
		dstr := base64.StdEncoding.EncodeToString(d)
		dmap[dstr] = true
	}
	for _, d := range deps {
		dstr := base64.StdEncoding.EncodeToString(d.ContextFile.DataDepDig)
		if !dmap[dstr] {
			return errors.New("ContextFile's deps were not all in dep slice")
		}
	}
	return nil
}

//DoesDependOn check if dig is contained in ContextFile DependOn
func DoesDependOn(dig []byte, l *ContextFile) bool {
	for _, dd := range l.DependsOn {
		if bytes.Equal(dd, dig) {
			return true
		}
	}
	return false
}
