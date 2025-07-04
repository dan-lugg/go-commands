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
	t.Run("default", func(t *testing.T) {
		mappingCatalog := commands.NewMappingCatalog()
		handlerCatalog := commands.NewHandlerCatalog()
		specWriter := NewSpecWriter(mappingCatalog, handlerCatalog)
		assert.Equal(t, specWriter.mappingCatalog, mappingCatalog)
		assert.Equal(t, specWriter.handlerCatalog, handlerCatalog)
	})

	t.Run("with options", func(t *testing.T) {
		mappingCatalog := commands.NewMappingCatalog()
		handlerCatalog := commands.NewHandlerCatalog()
		specWriter := NewSpecWriter(mappingCatalog, handlerCatalog,
			WithTitle("Test API"),
			WithVersion("2.0.0"),
			WithDescription("Test API for handling commands"))
		assert.Equal(t, specWriter.mappingCatalog, mappingCatalog)
		assert.Equal(t, specWriter.handlerCatalog, handlerCatalog)
		assert.Equal(t, specWriter.title, "Test API")
		assert.Equal(t, specWriter.version, "2.0.0")
		assert.Equal(t, specWriter.description, "Test API for handling commands")
	})
}

func TestSpecWriter_CreatePathItem(t *testing.T) {
	mappingCatalog := commands.NewMappingCatalog()
	handlerCatalog := commands.NewHandlerCatalog()
	specWriter := NewSpecWriter(mappingCatalog, handlerCatalog)
	reqType := reflect.TypeFor[AddCommandReq]()
	resType := reflect.TypeFor[AddCommandRes]()
	pathItem, err := specWriter.CreatePathItem("add", reqType, resType)
	assert.NoError(t, err)
	assert.NotNil(t, pathItem)
}

func TestSpecWriter_WriteSpec(t *testing.T) {
	const ExpectSpec = `{"info":{"description":"API for handling commands","title":"Commands API","version":"1.0.0"},"openapi":"3.0.0","paths":{"/add":{"post":{"description":"Handles the add command","operationId":"add","requestBody":{"content":{"application/json":{"schema":{"properties":{"argX":{"$ref":"int"},"argY":{"$ref":"int"}},"type":"object"}}},"required":true},"responses":{"200":{"content":{"application/json":{"schema":{"properties":{"result":{"$ref":"int"}},"type":"object"}}}},"default":{"description":""}},"summary":"HandleRaw add"}},"/sub":{"post":{"description":"Handles the sub command","operationId":"sub","requestBody":{"content":{"application/json":{"schema":{"properties":{"argX":{"$ref":"int"},"argY":{"$ref":"int"}},"type":"object"}}},"required":true},"responses":{"200":{"content":{"application/json":{"schema":{"properties":{"result":{"$ref":"int"}},"type":"object"}}}},"default":{"description":""}},"summary":"HandleRaw sub"}}}}`
	mappingCatalog := commands.NewMappingCatalog()
	handlerCatalog := commands.NewHandlerCatalog()
	specWriter := NewSpecWriter(mappingCatalog, handlerCatalog)
	commands.InsertHandler[AddCommandReq, AddCommandRes](handlerCatalog, func() commands.Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})
	commands.InsertHandler[SubCommandReq, SubCommandRes](handlerCatalog, func() commands.Handler[SubCommandReq, SubCommandRes] {
		return &SubHandler{}
	})
	commands.InsertMapping[AddCommandReq](mappingCatalog, AddReqName)
	commands.InsertMapping[SubCommandReq](mappingCatalog, SubReqName)
	data := []byte{}
	buffer := bytes.NewBuffer(data)
	writer := io.Writer(buffer)
	err := specWriter.WriteSpec(writer)
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(ExpectSpec), strings.TrimSpace(buffer.String()))
}

// </editor-fold>
