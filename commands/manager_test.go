package commands

import (
	"context"
	"github.com/dan-lugg/go-commands/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewManager(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		mappingCatalog := NewMappingCatalog()
		decoderCatalog := NewDecoderCatalog()
		handlerCatalog := NewHandlerCatalog()
		manager := NewManager(mappingCatalog, decoderCatalog, handlerCatalog)
		assert.NotNil(t, manager)
		assert.IsType(t, &Manager{}, manager)
		assert.Empty(t, manager.mappingCatalog.nameMappings)
		assert.Empty(t, manager.decoderCatalog.decoders)
		assert.Empty(t, manager.handlerCatalog.adapters)
	})

	t.Run("with options", func(t *testing.T) {
		mappingCatalog := NewMappingCatalog()
		decoderCatalog := NewDecoderCatalog()
		handlerCatalog := NewHandlerCatalog()
		manager := NewManager(mappingCatalog, decoderCatalog, handlerCatalog, func(*Manager) {})
		assert.NotNil(t, manager)
		assert.IsType(t, &Manager{}, manager)
		assert.Empty(t, manager.mappingCatalog.nameMappings)
		assert.Empty(t, manager.decoderCatalog.decoders)
		assert.Empty(t, manager.handlerCatalog.adapters)
	})
}

func TestInsert(t *testing.T) {
	mappingCatalog := NewMappingCatalog()
	decoderCatalog := NewDecoderCatalog()
	handlerCatalog := NewHandlerCatalog()
	manager := NewManager(mappingCatalog, decoderCatalog, handlerCatalog)

	t.Run("insert", func(t *testing.T) {
		Insert[AddCommandReq, AddCommandRes](manager, AddReqName, DefaultDecoder[AddCommandReq](), func() Handler[AddCommandReq, AddCommandRes] {
			return &AddHandler{}
		})
		assert.NotEmpty(t, manager.mappingCatalog.nameMappings)
		assert.Contains(t, manager.mappingCatalog.nameMappings, AddReqName)
	})
}

func Test_Manager_HandleRaw(t *testing.T) {
	mappingCatalog := NewMappingCatalog()
	decoderCatalog := NewDecoderCatalog()
	handlerCatalog := NewHandlerCatalog()
	manager := NewManager(mappingCatalog, decoderCatalog, handlerCatalog)
	Insert[AddCommandReq, AddCommandRes](manager, AddReqName, DefaultDecoder[AddCommandReq](), func() Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})

	t.Run("valid request", func(t *testing.T) {
		res, err := manager.HandleRaw(AddReqName, []byte(`{"argX": 3, "argY": 4}`), context.Background())
		assert.NoError(t, err)
		assert.Equal(t, AddCommandRes{Result: 7}, res)
	})

	t.Run("invalid request", func(t *testing.T) {
		res, err := manager.HandleRaw(SubReqName, []byte(`{"argX": 3, "argY": 4}`), context.Background())
		assert.Error(t, err)
		assert.ErrorIs(t, err, util.ErrNotCataloged)
		assert.Nil(t, res)
	})
}

func Test_HandleRaw(t *testing.T) {
	mappingCatalog := NewMappingCatalog()
	decoderCatalog := NewDecoderCatalog()
	handlerCatalog := NewHandlerCatalog()
	manager := NewManager(mappingCatalog, decoderCatalog, handlerCatalog)
	Insert[AddCommandReq, AddCommandRes](manager, AddReqName, DefaultDecoder[AddCommandReq](), func() Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})

	t.Run("valid request", func(t *testing.T) {
		res, err := HandleRaw(manager, AddReqName, []byte(`{"argX": 3, "argY": 4}`), context.Background())
		assert.NoError(t, err)
		assert.Equal(t, AddCommandRes{Result: 7}, res)
	})

	t.Run("invalid request", func(t *testing.T) {
		res, err := HandleRaw(manager, SubReqName, []byte(`{"argX": 3, "argY": 4}`), context.Background())
		assert.Error(t, err)
		assert.ErrorIs(t, err, util.ErrNotCataloged)
		assert.Nil(t, res)
	})
}

func Test_Manager_HandleReq(t *testing.T) {
	mappingCatalog := NewMappingCatalog()
	decoderCatalog := NewDecoderCatalog()
	handlerCatalog := NewHandlerCatalog()
	manager := NewManager(mappingCatalog, decoderCatalog, handlerCatalog)
	Insert[AddCommandReq, AddCommandRes](manager, AddReqName, DefaultDecoder[AddCommandReq](), func() Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})

	t.Run("valid request", func(t *testing.T) {
		res, err := manager.HandleReq(AddCommandReq{ArgX: 3, ArgY: 4}, context.Background())
		assert.NoError(t, err)
		assert.Equal(t, AddCommandRes{Result: 7}, res)
	})

	t.Run("invalid request", func(t *testing.T) {
		res, err := manager.HandleReq(SubCommandReq{ArgX: 3, ArgY: 4}, context.Background())
		assert.Error(t, err)
		assert.ErrorIs(t, err, util.ErrNotCataloged)
		assert.Nil(t, res)
	})
}

func Test_HandleReq(t *testing.T) {
	mappingCatalog := NewMappingCatalog()
	decoderCatalog := NewDecoderCatalog()
	handlerCatalog := NewHandlerCatalog()
	manager := NewManager(mappingCatalog, decoderCatalog, handlerCatalog)
	Insert[AddCommandReq, AddCommandRes](manager, AddReqName, DefaultDecoder[AddCommandReq](), func() Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})

	t.Run("valid request", func(t *testing.T) {
		res, err := HandleReq[AddCommandReq, AddCommandRes](manager, AddCommandReq{ArgX: 3, ArgY: 4}, context.Background())
		assert.NoError(t, err)
		assert.Equal(t, AddCommandRes{Result: 7}, res)
	})

	t.Run("invalid request", func(t *testing.T) {
		res, err := HandleReq[SubCommandReq, SubCommandRes](manager, SubCommandReq{ArgX: 3, ArgY: 4}, context.Background())
		assert.Error(t, err)
		assert.ErrorIs(t, err, util.ErrNotCataloged)
		assert.Zero(t, res)
	})
}
