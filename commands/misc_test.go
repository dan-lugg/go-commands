package commands

import (
	"context"
	"time"
)

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

func (h *AddHandler) Handle(ctx context.Context, req AddCommandReq) (res AddCommandRes, err error) {
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

func (h *SubHandler) Handle(ctx context.Context, req SubCommandReq) (res SubCommandRes, err error) {
	result := req.ArgX - req.ArgY
	return SubCommandRes{Result: result}, nil
}

type SlowCommandRes struct {
	CommandRes
	Name string
}

type SlowCommandReq struct {
	CommandReq[SlowCommandRes]
	Name string
	Fail bool
	Iter int
}

type SlowHandler struct {
	Handler[SlowCommandReq, SlowCommandRes]
}

func (h *SlowHandler) Handle(ctx context.Context, req SlowCommandReq) (res SlowCommandRes, err error) {
	for i := 1; i <= req.Iter; i++ {
		time.Sleep(100 * time.Millisecond)
		if ctx.Err() != nil {
			return SlowCommandRes{}, ctx.Err()
		}
	}
	return SlowCommandRes{
		Name: req.Name,
	}, nil
}
