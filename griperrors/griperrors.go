package griperrors

const (
	//EnUs US English
	EnUs = "en-us"
	//EsMx Espanol Mexicano
	EsMx = "es-mx"
	DEF  = EnUs
)

var ShareNodeKeyEmpty error = GErr(2).Msg(EnUs, "Key cannot be empty")
var WrongType error = GErr(3).Msg(EnUs, "Wrong type")
var TargetNodeNil error = GErr(5).Msg(EnUs, "Target node id cannot be nil")
var AccountNotEnabled error = GErr(6).Msg(EnUs, "Node account is not enabled")
var NodeAccountKeyNotFound error = GErr(7).Msg(EnUs, "Node account key not found")
var NodeAccountKeyUsed error = GErr(8).Msg(EnUs, "Onetime node account key has been used")
var NodeAccountKeyExpired error = GErr(9).Msg(EnUs, "Node account key has expired")
var AccountNotFound error = GErr(10).Msg(EnUs, "Account could not be found")
var AccountKeyNotAllowed error = GErr(11).Msg(EnUs, "Account does not allow node account keys")
var MaxNodesForAccount error = GErr(12).Msg(EnUs, "Maximum number of nodes for account reached")
var NotContextSource error = GErr(13).Msg(EnUs, "You don't have permission to add files to context")
var DependencyProblems error = GErr(14).Msg(EnUs, "Context file dependency problems found")
var InvalidFileSize error = GErr(15).Msg(EnUs, "ContextFile had invalid file size")

func GErr(code int) *Griperr {
	var g Griperr
	g.Code = code
	g.Message = make(map[string]string)
	return &g
}

type Griperr struct {
	Code    int
	Message map[string]string
}

func (e *Griperr) Msg(l string, m string) *Griperr {
	e.Message[l] = m
	return e
}

func (e *Griperr) Error() string {
	return e.Message[DEF]
}
