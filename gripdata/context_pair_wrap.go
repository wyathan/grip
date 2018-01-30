package gripdata

import (
	"encoding/base64"
)

type ContextPairWrap struct {
	Request  *ContextRequest
	Response *ContextResponse
}

func (p *ContextPairWrap) IsFullRepo() bool {
	return p.Request.FullRepo && p.Response.FullRepo
}

func (p *ContextPairWrap) GetIdStr() string {
	return base64.StdEncoding.EncodeToString(p.Request.ContextDig)
}
