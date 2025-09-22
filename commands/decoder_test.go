package commands

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewDecoderCatalog(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		catalog := NewDefaultDecoderCatalog()
		assert.NotNil(t, catalog)
		assert.Empty(t, catalog.decoders)
		assert.IsType(t, &DefaultDecoderCatalog{}, catalog)
	})

	t.Run("with options", func(t *testing.T) {
		catalog := NewDefaultDecoderCatalog(func(*DefaultDecoderCatalog) {})
		assert.NotNil(t, catalog)
		assert.Empty(t, catalog.decoders)
		assert.IsType(t, &DefaultDecoderCatalog{}, catalog)
	})
}

func Test_DecoderCatalog_Insert(t *testing.T) {
	t.Run("empty catalog", func(t *testing.T) {
		catalog := DefaultDecoderCatalog{}
		assert.Nil(t, catalog.decoders)
		catalog.Insert(reflect.TypeFor[AddCommandReq](), DefaultDecoder[AddCommandReq]())
		assert.NotEmpty(t, catalog.decoders)
		assert.Contains(t, catalog.decoders, reflect.TypeFor[AddCommandReq]())
	})

	t.Run("constructed catalog", func(t *testing.T) {
		catalog := NewDefaultDecoderCatalog()
		assert.NotNil(t, catalog)
		catalog.Insert(reflect.TypeFor[AddCommandReq](), DefaultDecoder[AddCommandReq]())
		assert.NotEmpty(t, catalog.decoders)
		assert.Contains(t, catalog.decoders, reflect.TypeFor[AddCommandReq]())
	})
}

func Test_InsertDecoder(t *testing.T) {
	catalog := NewDefaultDecoderCatalog()
	InsertDecoder[AddCommandReq](catalog, DefaultDecoder[AddCommandReq]())
	assert.NotEmpty(t, catalog.decoders)
	assert.Contains(t, catalog.decoders, reflect.TypeFor[AddCommandReq]())
}

func Test_DecoderCatalog_Decode(t *testing.T) {
	catalog := NewDefaultDecoderCatalog()
	InsertDecoder[AddCommandReq](catalog, DefaultDecoder[AddCommandReq]())

	t.Run("valid input", func(t *testing.T) {
		req, err := catalog.Decode(reflect.TypeFor[AddCommandReq](), []byte(`{"argX": 3, "argY": 4}`))
		assert.NoError(t, err)
		assert.Equal(t, AddCommandReq{ArgX: 3, ArgY: 4}, req)
	})

	t.Run("empty input", func(t *testing.T) {
		req, err := catalog.Decode(reflect.TypeFor[AddCommandReq](), []byte(`{}`))
		assert.NoError(t, err)
		assert.Equal(t, AddCommandReq{}, req)
	})

	t.Run("invalid input", func(t *testing.T) {
		req, err := catalog.Decode(reflect.TypeFor[AddCommandReq](), []byte(`#!`))
		assert.Error(t, err)
		assert.Nil(t, req)
	})

	t.Run("decoder missing", func(t *testing.T) {
		req, err := catalog.Decode(reflect.TypeFor[SubCommandReq](), []byte(`#!`))
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrDecoderMissing)
		assert.Nil(t, req)
	})
}

func Test_DefaultCommandReqDecoder(t *testing.T) {
	decoder := DefaultDecoder[AddCommandReq]()

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
