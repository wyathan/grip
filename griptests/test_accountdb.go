package griptests

import (
	"encoding/base64"
	"fmt"

	"github.com/wyathan/grip/gripdata"
)

/*
	StoreAccount(a *gripdata.Account) error
	GetAccount(s string) *gripdata.Account
	StoreNodeAccountKey(ak *gripdata.NodeAccountKey) error
	GetNodeAccountKey(key string) *gripdata.NodeAccountKey
	StoreNodeAccount(na *gripdata.NodeAccount) error
	GetNodeAccount(id []byte) *gripdata.NodeAccount
	IncrNumberContexts(a *gripdata.Account) error
	IncrNumberNodes(a *gripdata.Account) error
	CheckUpdateStorageUsed(a *gripdata.Account, fsize uint64) error
*/

func (t *TestDB) StoreAccount(a *gripdata.Account) error {
	t.Lock()
	defer t.Unlock()
	if t.Accounts == nil {
		t.Accounts = make(map[string]*gripdata.Account)
	}
	t.Accounts[a.AccountID] = a
	return nil
}
func (t *TestDB) GetAccount(s string) *gripdata.Account {
	t.Lock()
	defer t.Unlock()
	return t.Accounts[s]
}
func (t *TestDB) StoreNodeAccountKey(ak *gripdata.NodeAccountKey) error {
	t.Lock()
	defer t.Unlock()
	if t.AccountKeys == nil {
		t.AccountKeys = make(map[string]*gripdata.NodeAccountKey)
	}
	t.AccountKeys[ak.Key] = ak
	return nil
}
func (t *TestDB) GetNodeAccountKey(key string) *gripdata.NodeAccountKey {
	t.Lock()
	defer t.Unlock()
	return t.AccountKeys[key]
}
func (t *TestDB) StoreNodeAccount(na *gripdata.NodeAccount) error {
	t.Lock()
	defer t.Unlock()
	if t.NodeAccounts == nil {
		t.NodeAccounts = make(map[string]*gripdata.NodeAccount)
	}
	t.NodeAccounts[base64.StdEncoding.EncodeToString(na.NodeID)] = na
	return nil
}
func (t *TestDB) GetNodeAccount(id []byte) *gripdata.NodeAccount {
	t.Lock()
	defer t.Unlock()
	return t.NodeAccounts[base64.StdEncoding.EncodeToString(id)]
}
func (t *TestDB) IncrNumberContexts(a *gripdata.Account) error {
	t.Lock()
	defer t.Unlock()
	ua := t.Accounts[a.AccountID]
	ua.NumberContexts++
	return nil
}
func (t *TestDB) IncrNumberNodes(a *gripdata.Account) error {
	t.Lock()
	defer t.Unlock()
	ua := t.Accounts[a.AccountID]
	ua.NumberNodes++
	return nil
}
func (t *TestDB) CheckUpdateStorageUsed(a *gripdata.Account, fsize uint64) error {
	t.Lock()
	defer t.Unlock()
	a = t.Accounts[a.AccountID]
	if a.DiskSpaceUsed+fsize > a.MaxDiskSpace {
		return fmt.Errorf("would exceed max storage %d >= %d", a.DiskSpaceUsed+fsize, a.MaxDiskSpace)
	}
	a.DiskSpaceUsed = a.DiskSpaceUsed + fsize
	t.Accounts[a.AccountID] = a
	return nil
}
