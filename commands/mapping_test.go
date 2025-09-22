package commands

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewMappingCatalog(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		catalog := NewMappingCatalog()
		assert.NotNil(t, catalog)
		assert.Empty(t, catalog.nameMappings)
		assert.IsType(t, &DefaultMappingCatalog{}, catalog)
	})

	t.Run("with options", func(t *testing.T) {
		catalog := NewMappingCatalog(func(*DefaultMappingCatalog) {})
		assert.NotNil(t, catalog)
		assert.Empty(t, catalog.nameMappings)
		assert.IsType(t, &DefaultMappingCatalog{}, catalog)
	})
}

func Test_MappingCatalog_Insert(t *testing.T) {
	t.Run("empty catalog", func(t *testing.T) {
		catalog := DefaultMappingCatalog{}
		assert.Nil(t, catalog.nameMappings)
		catalog.Insert(AddReqName, reflect.TypeFor[AddCommandReq]())
		assert.NotEmpty(t, catalog.nameMappings)
		assert.Contains(t, catalog.nameMappings, AddReqName)
	})

	t.Run("constructed catalog", func(t *testing.T) {
		catalog := NewMappingCatalog()
		assert.NotNil(t, catalog)
		catalog.Insert(AddReqName, reflect.TypeFor[AddCommandReq]())
		assert.NotEmpty(t, catalog.nameMappings)
		assert.Contains(t, catalog.nameMappings, AddReqName)
	})
}

func Test_MappingCatalog_ByName(t *testing.T) {
	catalog := NewMappingCatalog()
	InsertMapping[AddCommandReq](catalog, AddReqName)

	t.Run("default", func(t *testing.T) {
		reqType, err := catalog.ByName(AddReqName)
		assert.NoError(t, err)
		assert.Equal(t, reflect.TypeFor[AddCommandReq](), reqType)
	})

	t.Run("mapping missing", func(t *testing.T) {
		reqType, err := catalog.ByName(SubReqName)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrMappingMissing)
		assert.Nil(t, reqType)
	})
}

func Test_MappingCatalog_ByType(t *testing.T) {
	catalog := NewMappingCatalog()
	InsertMapping[AddCommandReq](catalog, AddReqName)

	t.Run("default", func(t *testing.T) {
		reqType, err := catalog.ByType(reflect.TypeFor[AddCommandReq]())
		assert.NoError(t, err)
		assert.Equal(t, AddReqName, reqType)
	})

	t.Run("mapping missing", func(t *testing.T) {
		reqType, err := catalog.ByType(reflect.TypeFor[SubCommandReq]())
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrMappingMissing)
		assert.Empty(t, reqType)
	})
}

func Test_InsertMapping(t *testing.T) {
}
