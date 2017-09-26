package gripdata

//NodeAccount Accociates a node with an acocunt
//NOTE: This data is local only and never shared with other nodes.
type NodeAccount struct {
	NodeID    []byte //The id of the node that belongs to the account
	AccountID string //The account id
	MetaData  string //Generic metadata string
	Enabled   bool   //Does the owner of this account want this node to be enable
}
