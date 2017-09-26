package gripdata

//Account is just that.  It limits the permissions of nodes
//accociated with this account on this node.
//NOTE: This data is local only and never shared with other nodes.
type Account struct {
	AccountID          string //Userid
	AuthenticationData []byte //Probably just password hash.
	//NOTE: This password does not allow context participation

	Email    string //Email if needed/provided
	MetaData string //Generic metadata

	MaxNodes            uint32 //Maximum number of nodes that can associate with this account
	MaxDiskSpace        uint64 //Maximum disk space this account can use
	AllowFullRepo       bool   //Should this node keep a full snapshot of contexts when requested
	AllowContextSource  bool   //Allow creating context data from this node
	AllowContextNode    bool   //Allow adding new nodes to contexts
	AllowNodeAcocuntKey bool   //Allow use of NodeAccountKeys to assocate with this account
	AllowCacheMode      uint32 //Which cache modes are available

	Message string //Message to present to to the account user
	Enabled bool   //Is this account enabled

	NumberNodes   uint32 //Current number of nodes associated with this account
	DiskSpaceUsed uint64 //Current disk space used
}
