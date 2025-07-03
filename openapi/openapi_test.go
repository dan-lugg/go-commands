package openapi

import (
	"bytes"
	"context"
	"github.com/dan-lugg/go-commands/commands"
	"github.com/stretchr/testify/assert"
	"io"
	"reflect"
	"strings"
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
	commands.Handler[AddCommandReq, AddCommandRes]
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
	commands.Handler[SubCommandReq, SubCommandRes]
}

func (h *SubHandler) Handle(req SubCommandReq, ctx context.Context) (res SubCommandRes, err error) {
	result := req.ArgX - req.ArgY
	return SubCommandRes{Result: result}, nil
}

// </editor-fold>

// <editor-fold desc="Tests">

func TestNewSpecWriter(t *testing.T) {
	mappingRegistry := commands.NewMappingRegistry()
	handlerRegistry := commands.NewHandlerRegistry()
	specWriter := NewSpecWriter(mappingRegistry, handlerRegistry)
	assert.Equal(t, specWriter.mappingRegistry, mappingRegistry)
	assert.Equal(t, specWriter.handlerRegistry, handlerRegistry)
}

func TestSpecWriter_CreatePathItem(t *testing.T) {
	mappingRegistry := commands.NewMappingRegistry()
	handlerRegistry := commands.NewHandlerRegistry()
	specWriter := NewSpecWriter(mappingRegistry, handlerRegistry)
	reqType := reflect.TypeFor[AddCommandReq]()
	resType := reflect.TypeFor[AddCommandRes]()
	pathItem, err := specWriter.CreatePathItem("add", reqType, resType)
	assert.NoError(t, err)
	assert.NotNil(t, pathItem)
}

func TestSpecWriter_WriteSpec(t *testing.T) {
	const ExpectSpec = `{"/add":{"post":{"description":"Handles the add command","operationId":"add","requestBody":{"content":{"application/json":{"schema":{"properties":{"argX":{"$ref":"int"},"argY":{"$ref":"int"}},"type":"object"}}},"required":true},"responses":{"200":{"content":{"application/json":{"schema":{"properties":{"result":{"$ref":"int"}},"type":"object"}}}},"default":{"description":""}},"summary":"Handle add"}},"/sub":{"post":{"description":"Handles the sub command","operationId":"sub","requestBody":{"content":{"application/json":{"schema":{"properties":{"argX":{"$ref":"int"},"argY":{"$ref":"int"}},"type":"object"}}},"required":true},"responses":{"200":{"content":{"application/json":{"schema":{"properties":{"result":{"$ref":"int"}},"type":"object"}}}},"default":{"description":""}},"summary":"Handle sub"}}}`

	mappingRegistry := commands.NewMappingRegistry()
	handlerRegistry := commands.NewHandlerRegistry()
	specWriter := NewSpecWriter(mappingRegistry, handlerRegistry)

	commands.RegisterHandler[AddCommandReq, AddCommandRes](handlerRegistry, func() commands.Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})
	commands.RegisterHandler[SubCommandReq, SubCommandRes](handlerRegistry, func() commands.Handler[SubCommandReq, SubCommandRes] {
		return &SubHandler{}
	})

	commands.RegisterMapping[AddCommandReq](mappingRegistry, AddReqName)
	commands.RegisterMapping[SubCommandReq](mappingRegistry, SubReqName)
	data := []byte{}
	buffer := bytes.NewBuffer(data)
	writer := io.Writer(buffer)

	err := specWriter.WriteSpec(writer)
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(ExpectSpec), strings.TrimSpace(buffer.String()))
}

// </editor-fold>
