package gripdata

//MyNodePrivateData is private information about this node
//never shared with another.
type MyNodePrivateData struct {
	ID                []byte //Unique id for the node (digest of public key)
	BindAddress       string //where to bind listen port
	BindPort          uint32 //listen port to use
	PrivateKey        []byte
	PrivateMetaData   string //Extra private generic metadata
	AutoShareNodeInfo bool
	AutoShareMetaData string

	AutoCreateShareAccount         bool   //When we share info with another node, automatically grant some auth
	AutoAccountMaxNodes            uint32 //Maximum number of nodes that can associate with this account
	AutoAccountMaxDiskSpace        uint64 //Maximum disk space this account can use
	AutoAccountAllowFullRepo       bool   //Should this node keep a full snapshot of contexts when requested
	AutoAccountAllowContextSource  bool   //Allow creating context data from this node
	AutoAccountAllowContextNode    bool   //Allow adding new nodes to contexts
	AutoAccountAllowNodeAcocuntKey bool   //Allow use of NodeAccountKeys to assocate with this account
	AutoAccountAllowCacheMode      uint32 //Which cache modes are available

	AutoContextResponse bool //Automatically reply to context requests
	//The account data specifies how we respond to a node's request

}
