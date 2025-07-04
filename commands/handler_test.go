package commands

import (
	"context"
	"github.com/dan-lugg/go-commands/util"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

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

func Test_NewHandlerCatalog(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		catalog := NewHandlerCatalog()
		assert.NotNil(t, catalog)
		assert.Empty(t, catalog.adapters)
		assert.IsType(t, &HandlerCatalog{}, catalog)
	})

	t.Run("with options", func(t *testing.T) {
		catalog := NewHandlerCatalog(func(*HandlerCatalog) {})
		assert.NotNil(t, catalog)
		assert.Empty(t, catalog.adapters)
		assert.IsType(t, &HandlerCatalog{}, catalog)
	})
}

func Test_HandlerCatalog_Insert(t *testing.T) {
	t.Run("empty catalog", func(t *testing.T) {
		catalog := HandlerCatalog{}
		assert.Nil(t, catalog.adapters)
		adapter := NewDefaultHandlerAdapter(func() Handler[AddCommandReq, AddCommandRes] {
			return &AddHandler{}
		})
		catalog.Insert(adapter)
		assert.NotEmpty(t, catalog.adapters)
		assert.Contains(t, catalog.adapters, adapter.ReqType())
	})

	t.Run("constructed catalog", func(t *testing.T) {
		catalog := NewHandlerCatalog()
		assert.NotNil(t, catalog)
		adapter := NewDefaultHandlerAdapter(func() Handler[AddCommandReq, AddCommandRes] {
			return &AddHandler{}
		})
		catalog.Insert(adapter)
		assert.NotEmpty(t, catalog.adapters)
		assert.Contains(t, catalog.adapters, adapter.ReqType())
	})
}

func Test_InsertHandler(t *testing.T) {
	catalog := NewHandlerCatalog()
	InsertHandler[AddCommandReq, AddCommandRes](catalog, func() Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})
	assert.NotEmpty(t, catalog.adapters)
	assert.Contains(t, catalog.adapters, reflect.TypeFor[AddCommandReq]())
}

func Test_HandlerCatalog_Handle(t *testing.T) {
	catalog := NewHandlerCatalog()
	InsertHandler[AddCommandReq, AddCommandRes](catalog, func() Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})

	t.Run("req type handler cataloged", func(t *testing.T) {
		res, err := catalog.Handle(AddCommandReq{ArgX: 3, ArgY: 4}, context.Background())
		assert.NoError(t, err)
		assert.Equal(t, AddCommandRes{Result: 7}, res)
	})

	t.Run("req type handler not cataloged", func(t *testing.T) {
		res, err := catalog.Handle(SubCommandReq{ArgX: 3, ArgY: 4}, context.Background())
		assert.Error(t, err)
		assert.ErrorIs(t, err, util.ErrNotCataloged)
		assert.Nil(t, res)
	})
}

func Test_HandlerCatalog_TypeMap(t *testing.T) {
	catalog := NewHandlerCatalog()
	InsertHandler[AddCommandReq, AddCommandRes](catalog, func() Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})
	InsertHandler[SubCommandReq, SubCommandRes](catalog, func() Handler[SubCommandReq, SubCommandRes] {
		return &SubHandler{}
	})
	typeMap := catalog.TypeMap()
	assert.NotEmpty(t, typeMap)
	assert.Contains(t, typeMap, reflect.TypeFor[AddCommandReq]())
	assert.Contains(t, typeMap, reflect.TypeFor[SubCommandReq]())
	assert.Equal(t, reflect.TypeFor[AddCommandRes](), typeMap[reflect.TypeFor[AddCommandReq]()])
	assert.Equal(t, reflect.TypeFor[SubCommandRes](), typeMap[reflect.TypeFor[SubCommandReq]()])
}
