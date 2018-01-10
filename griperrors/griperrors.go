package griperrors

const (
	//EnUs US English
	EnUs = "en-us"
	//EsMx Espanol Mexicano
	EsMx = "es-mx"
	DEF  = EnUs
)

var ShareNodeKeyEmpty error = GErr(2).Msg(EnUs, "Key cannot be empty")
var NotNode error = GErr(3).Msg(EnUs, "Not a node")
var NotShareNode error = GErr(4).Msg(EnUs, "Not ShareNodeInfo")
var TargetNodeNil error = GErr(5).Msg(EnUs, "Target node id cannot be nil")
var AccountNotEnabled error = GErr(6).Msg(EnUs, "Node account is not enabled")

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
