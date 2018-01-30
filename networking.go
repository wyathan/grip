package grip

import (
	"time"

	"github.com/wyathan/grip/gripdata"
)

//MAXSEND The maximum number of SendData to query from
//the database and send on the connection at one time
const MAXSEND int = 10

//SLEEPONNOSEND The number of seconds to sleep between
//querying the database for new data to send
const SLEEPONNOSEND time.Duration = 1 * time.Second

//BUFFERSIZE is the send channel size
//Note: The test network sometimes results in a deadlock where
//the ReadRoutine will block because the SendChan is full.
//The WriteRoutine won't clear SendChan because it's blocked
//trying to add to the testnetwork Send channel.  It is full because
//the other end of the connection's ReadRoutine is also blocked, etc.
//Increasing this value mitigates this issue.
const BUFFERSIZE int = 20

//MAXCONNECTIONS is the maximum number of connections we allow
const MAXCONNECTIONS int = 1000

//MAXCONNECTIONATTEMPTS number of connections to attempt at once
const MAXCONNECTIONATTEMPTS int = 20

//TRYCONNECTAGAINAFTERFAIL how long to wait until we try to connect
//again after we've failed
const TRYCONNECTAGAINAFTERFAIL time.Duration = 1 * time.Second

//WAITUNTILCONNECTAGAIN give connections this amount of time to
//become established and upated in the database before we try
//to connect again.  It should take less time than this for both
//ends of a connection to close after one side closes, otherwise
//existing connections that haven't yet completely closed will
//keep new connections from being made.
const WAITUNTILCONNECTAGAIN time.Duration = TRYCONNECTAGAINAFTERFAIL

//CONNECTROUTINESLEEP is how long the connect routine sleeps
//before it trys to make new connections
const CONNECTROUTINESLEEP time.Duration = 500 * time.Millisecond

//Connection handles sending and reading data
//on a connection
type Connection interface {
	Read() (interface{}, error)
	Send(d interface{}) error
	GetNodeID() []byte //The validated id of the node connected to
	Close()            //Must cause blocking read to immediately exit
}

//Socket handles making connections and accepting them
type Socket interface {
	ConnectTo(n *gripdata.Node) (Connection, error)
	Accept() (Connection, error)
	Close()
}

//CheckDig check if a node has a digest you want to send it
type CheckDig struct {
	Dig []byte
}

//RespDig lets other node know if we have this object yet or not
type RespDig struct {
	HaveIt bool
	Dig    []byte
}

//SendDig actually send this data
type SendDig struct {
	HaveIt bool
	Dig    []byte
}

//RejectDig indicates this digest was rejected by the receving node
type RejectDig struct {
	Dig     []byte
	Message string
}

//AckDig indicates the receiving node got the data
type AckDig struct {
	Dig []byte
}

//ReqContextFile requests a ContextFile be (re)sent
//to this node.  Most likely so that it can just
//send it along to another node.
type ReqContextFile struct {
	Dig []byte
}
