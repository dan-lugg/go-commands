package commands

import (
	"context"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

// <editor-fold desc="Types">

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

// </editor-fold>

// <editor-fold desc="Tests">

func Test_NewDecoderRegistry(t *testing.T) {
}

func Test_DecoderRegistry_RegisterDecoder(t *testing.T) {
	registry := NewDecoderRegistry()

	registry.RegisterDecoder(AddReqName, reflect.TypeFor[AddCommandReq](), DefaultCommandReqDecoder[AddCommandReq]())
	assert.NotEmpty(t, registry.decoders)
	assert.Contains(t, registry.decoders, reflect.TypeFor[AddCommandReq]())
}

func Test_RegisterDecoder(t *testing.T) {
	registry := NewDecoderRegistry()

	RegisterDecoder[AddCommandReq](registry, AddReqName, DefaultCommandReqDecoder[AddCommandReq]())
	assert.NotEmpty(t, registry.decoders)
	assert.Contains(t, registry.decoders, reflect.TypeFor[AddCommandReq]())
}

func Test_DecoderRegistry_Decode(t *testing.T) {
	registry := NewDecoderRegistry()
	RegisterDecoder[AddCommandReq](registry, AddReqName, DefaultCommandReqDecoder[AddCommandReq]())

	t.Run("valid input", func(t *testing.T) {
		req, err := registry.Decode(AddReqName, []byte(`{"argX": 3, "argY": 4}`))
		assert.NoError(t, err)
		assert.Equal(t, AddCommandReq{ArgX: 3, ArgY: 4}, req)
	})

	t.Run("empty input", func(t *testing.T) {
		req, err := registry.Decode(AddReqName, []byte(`{}`))
		assert.NoError(t, err)
		assert.Equal(t, AddCommandReq{}, req)
	})

	t.Run("invalid input", func(t *testing.T) {
		req, err := registry.Decode(AddReqName, []byte(`#!`))
		assert.Error(t, err)
		assert.Nil(t, req)
	})
}

func Test_DefaultCommandReqDecoder(t *testing.T) {
	decoder := DefaultCommandReqDecoder[AddCommandReq]()

	t.Run("valid input", func(t *testing.T) {
		req, err := decoder([]byte(`{"argX": 3, "argY": 4}`))
		assert.NoError(t, err)
		assert.Equal(t, AddCommandReq{ArgX: 3, ArgY: 4}, req)
	})

	t.Run("empty input", func(t *testing.T) {
		req, err := decoder([]byte(`{}`))
		assert.NoError(t, err)
		assert.Equal(t, AddCommandReq{}, req)
	})

	t.Run("invalid input", func(t *testing.T) {
		req, err := decoder([]byte(`#!`))
		assert.Error(t, err)
		assert.Nil(t, req)
	})
}

func Test_NewDefaultHandlerAdapter(t *testing.T) {
	adapter := NewDefaultHandlerAdapter(func() Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})

	assert.NotNil(t, adapter)
	assert.IsType(t, &DefaultHandlerAdapter[AddCommandReq, AddCommandRes]{}, adapter)
}

func Test_DefaultHandlerAdapter_Handle(t *testing.T) {
	adapter := NewDefaultHandlerAdapter(func() Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})

	t.Run("valid req", func(t *testing.T) {
		res, err := adapter.Handle(AddCommandReq{ArgX: 3, ArgY: 4}, context.Background())
		assert.NoError(t, err)
		assert.Equal(t, AddCommandRes{Result: 7}, res)
	})

	t.Run("invalid req", func(t *testing.T) {
		res, err := adapter.Handle(SubCommandReq{}, context.Background())
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func Test_DefaultHandlerAdapter_ReqType(t *testing.T) {
	adapter := NewDefaultHandlerAdapter(func() Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})
	assert.Equal(t, reflect.TypeFor[AddCommandReq](), adapter.ReqType())
}

func Test_DefaultHandlerAdapter_ResType(t *testing.T) {
	adapter := NewDefaultHandlerAdapter(func() Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})
	assert.Equal(t, reflect.TypeFor[AddCommandRes](), adapter.ResType())
}

func Test_NewHandlerRegistry(t *testing.T) {
	registry := NewHandlerRegistry()

	assert.NotNil(t, registry)
	assert.Empty(t, registry.adapters)
}

func Test_HandlerRegistry_Register(t *testing.T) {
	registry := NewHandlerRegistry()
	adapter := NewDefaultHandlerAdapter(func() Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})
	registry.Register(adapter)

	assert.NotEmpty(t, registry.adapters)
	assert.Contains(t, registry.adapters, adapter.ReqType())
}

func Test_RegisterHandler(t *testing.T) {
	registry := NewHandlerRegistry()
	RegisterHandler[AddCommandReq, AddCommandRes](registry, func() Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})

	assert.NotEmpty(t, registry.adapters)
	assert.Contains(t, registry.adapters, reflect.TypeFor[AddCommandReq]())
}

func Test_HandlerRegistry_Handle(t *testing.T) {
	registry := NewHandlerRegistry()
	RegisterHandler[AddCommandReq, AddCommandRes](registry, func() Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})

	t.Run("registered req", func(t *testing.T) {
		res, err := registry.Handle(AddCommandReq{ArgX: 3, ArgY: 4}, context.Background())
		assert.NoError(t, err)
		assert.Equal(t, AddCommandRes{Result: 7}, res)
	})

	t.Run("unregistered req", func(t *testing.T) {
		res, err := registry.Handle(SubCommandReq{ArgX: 3, ArgY: 4}, context.Background())
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func Test_JSONResolver_Resolve(t *testing.T) {
}

// </editor-fold>
