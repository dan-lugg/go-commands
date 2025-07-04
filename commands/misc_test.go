package commands

import "context"

const (
	AddReqName = "add"
	SubReqName = "sub"
)

type AddCommandRes struct {
	Result int `json:"result"`
}

type AddCommandReq struct {
	ArgX int `json:"argX"`
	ArgY int `json:"argY"`
}

type AddHandler struct {
	Handler[AddCommandReq, AddCommandRes]
}

func (h *AddHandler) Handle(req AddCommandReq, ctx context.Context) (res AddCommandRes, err error) {
	result := req.ArgX + req.ArgY
	return AddCommandRes{Result: result}, nil
}

type SubCommandRes struct {
	Result int `json:"result"`
}

type SubCommandReq struct {
	ArgX int `json:"argX"`
	ArgY int `json:"argY"`
}

type SubHandler struct {
	Handler[SubCommandReq, SubCommandRes]
}

func (h *SubHandler) Handle(req SubCommandReq, ctx context.Context) (res SubCommandRes, err error) {
	result := req.ArgX - req.ArgY
	return SubCommandRes{Result: result}, nil
}
