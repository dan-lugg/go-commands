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

func Test_NewMappingRegistry(t *testing.T) {
	registry := NewMappingRegistry()
	assert.NotNil(t, registry)
	assert.Empty(t, registry.nameMappings)
	assert.IsType(t, &MappingRegistry{}, registry)
}

func Test_MappingRegistry_Register(t *testing.T) {
	t.Run("empty registry", func(t *testing.T) {
		registry := MappingRegistry{}
		assert.Nil(t, registry.nameMappings)
		registry.Register(AddReqName, reflect.TypeFor[AddCommandReq]())
		assert.NotEmpty(t, registry.nameMappings)
		assert.Contains(t, registry.nameMappings, AddReqName)
	})

	t.Run("constructed registry", func(t *testing.T) {
		registry := NewMappingRegistry()
		assert.NotNil(t, registry)
		registry.Register(AddReqName, reflect.TypeFor[AddCommandReq]())
		assert.NotEmpty(t, registry.nameMappings)
		assert.Contains(t, registry.nameMappings, AddReqName)
	})
}

func Test_MappingRegistry_ByName(t *testing.T) {
	registry := NewMappingRegistry()
	RegisterMapping[AddCommandReq](registry, AddReqName)

	t.Run("valid input", func(t *testing.T) {
		reqType, err := registry.ByName(AddReqName)
		assert.NoError(t, err)
		assert.Equal(t, reflect.TypeFor[AddCommandReq](), reqType)
	})

	t.Run("mapping not registered", func(t *testing.T) {
		reqType, err := registry.ByName(SubReqName)
		assert.Error(t, err)
		assert.Nil(t, reqType)
	})
}

func Test_MappingRegistry_ByType(t *testing.T) {
	registry := NewMappingRegistry()
	RegisterMapping[AddCommandReq](registry, AddReqName)

	t.Run("valid input", func(t *testing.T) {
		reqType, err := registry.ByType(reflect.TypeFor[AddCommandReq]())
		assert.NoError(t, err)
		assert.Equal(t, AddReqName, reqType)
	})

	t.Run("mapping not registered", func(t *testing.T) {
		reqType, err := registry.ByType(reflect.TypeFor[SubCommandReq]())
		assert.Error(t, err)
		assert.Empty(t, reqType)
	})
}

func Test_RegisterMapping(t *testing.T) {
}

func Test_NewDecoderRegistry(t *testing.T) {
}

func Test_DecoderRegistry_Register(t *testing.T) {
	t.Run("empty registry", func(t *testing.T) {
		registry := DecoderRegistry{}
		assert.Nil(t, registry.decoders)
		registry.Register(reflect.TypeFor[AddCommandReq](), DefaultCommandReqDecoder[AddCommandReq]())
		assert.NotEmpty(t, registry.decoders)
		assert.Contains(t, registry.decoders, reflect.TypeFor[AddCommandReq]())
	})

	t.Run("constructed registry", func(t *testing.T) {
		registry := NewDecoderRegistry()
		assert.NotNil(t, registry)
		registry.Register(reflect.TypeFor[AddCommandReq](), DefaultCommandReqDecoder[AddCommandReq]())
		assert.NotEmpty(t, registry.decoders)
		assert.Contains(t, registry.decoders, reflect.TypeFor[AddCommandReq]())
	})
}

func Test_RegisterDecoder(t *testing.T) {
	registry := NewDecoderRegistry()
	RegisterDecoder[AddCommandReq](registry, DefaultCommandReqDecoder[AddCommandReq]())
	assert.NotEmpty(t, registry.decoders)
	assert.Contains(t, registry.decoders, reflect.TypeFor[AddCommandReq]())
}

func Test_DecoderRegistry_Decode(t *testing.T) {
	registry := NewDecoderRegistry()
	RegisterDecoder[AddCommandReq](registry, DefaultCommandReqDecoder[AddCommandReq]())

	t.Run("valid input", func(t *testing.T) {
		req, err := registry.Decode(reflect.TypeFor[AddCommandReq](), []byte(`{"argX": 3, "argY": 4}`))
		assert.NoError(t, err)
		assert.Equal(t, AddCommandReq{ArgX: 3, ArgY: 4}, req)
	})

	t.Run("empty input", func(t *testing.T) {
		req, err := registry.Decode(reflect.TypeFor[AddCommandReq](), []byte(`{}`))
		assert.NoError(t, err)
		assert.Equal(t, AddCommandReq{}, req)
	})

	t.Run("invalid input", func(t *testing.T) {
		req, err := registry.Decode(reflect.TypeFor[AddCommandReq](), []byte(`#!`))
		assert.Error(t, err)
		assert.Nil(t, req)
	})

	t.Run("decoder not registered", func(t *testing.T) {
		req, err := registry.Decode(reflect.TypeFor[SubCommandReq](), []byte(`#!`))
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
	t.Run("empty registry", func(t *testing.T) {
		registry := HandlerRegistry{}
		assert.Nil(t, registry.adapters)
		adapter := NewDefaultHandlerAdapter(func() Handler[AddCommandReq, AddCommandRes] {
			return &AddHandler{}
		})
		registry.Register(adapter)
		assert.NotEmpty(t, registry.adapters)
		assert.Contains(t, registry.adapters, adapter.ReqType())
	})

	t.Run("constructed registry", func(t *testing.T) {
		registry := NewHandlerRegistry()
		assert.NotNil(t, registry)
		adapter := NewDefaultHandlerAdapter(func() Handler[AddCommandReq, AddCommandRes] {
			return &AddHandler{}
		})
		registry.Register(adapter)
		assert.NotEmpty(t, registry.adapters)
		assert.Contains(t, registry.adapters, adapter.ReqType())
	})
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

func Test_HandlerRegistry_TypeMap(t *testing.T) {
	registry := NewHandlerRegistry()
	RegisterHandler[AddCommandReq, AddCommandRes](registry, func() Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})
	RegisterHandler[SubCommandReq, SubCommandRes](registry, func() Handler[SubCommandReq, SubCommandRes] {
		return &SubHandler{}
	})
	typeMap := registry.TypeMap()
	assert.NotEmpty(t, typeMap)
	assert.Contains(t, typeMap, reflect.TypeFor[AddCommandReq]())
	assert.Contains(t, typeMap, reflect.TypeFor[SubCommandReq]())
	assert.Equal(t, reflect.TypeFor[AddCommandRes](), typeMap[reflect.TypeFor[AddCommandReq]()])
	assert.Equal(t, reflect.TypeFor[SubCommandRes](), typeMap[reflect.TypeFor[SubCommandReq]()])
}

// </editor-fold>
