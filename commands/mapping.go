package commands

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/dan-lugg/go-commands/util"
)

var (
	ErrMappingMissing = errors.New("mapping missing")
)

type MappingCatalog interface {
	Insert(reqName string, reqType reflect.Type)
	ByName(reqName string) (reqType reflect.Type, err error)
	ByType(reqType reflect.Type) (reqName string, err error)
}

// DefaultMappingCatalog is a catalog for managing mappings between request names and types.
//
// Fields:
//   - mutex: A sync.RWMutex used to ensure thread-safe access to the catalog.
//   - nameMappings: A map that associates request names (strings) with their corresponding reflect.Type.
//   - typeMappings: A map that associates reflect.Type with their corresponding request names (strings).
type DefaultMappingCatalog struct {
	mutex        sync.RWMutex
	nameMappings map[string]reflect.Type
	typeMappings map[reflect.Type]string
}

type NewMappingCatalogOption = util.Option[*DefaultMappingCatalog]

// NewMappingCatalog creates and returns a new instance of DefaultMappingCatalog.
//
// The catalog is initialized with:
//   - A sync.RWMutex for thread-safe access.
//   - nameMappings: A map associating request names (strings) with their corresponding reflect.Type.
//   - typeMappings: A map associating reflect.Type with their corresponding request names (strings).
//
// Returns:
//   - A pointer to a DefaultMappingCatalog instance.
func NewMappingCatalog(options ...NewMappingCatalogOption) (catalog *DefaultMappingCatalog) {
	catalog = &DefaultMappingCatalog{
		mutex:        sync.RWMutex{},
		nameMappings: make(map[string]reflect.Type),
		typeMappings: make(map[reflect.Type]string),
	}
	for _, option := range options {
		option(catalog)
	}
	return catalog
}

// Insert adds a mapping between a request name and its corresponding type.
//
// Parameters:
//   - reqName: A string representing the name of the request.
//   - reqType: A reflect.Type representing the type of the request.
func (m *DefaultMappingCatalog) Insert(reqName string, reqType reflect.Type) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.nameMappings == nil {
		m.nameMappings = make(map[string]reflect.Type)
	}
	if m.typeMappings == nil {
		m.typeMappings = make(map[reflect.Type]string)
	}
	m.nameMappings[reqName] = reqType
	m.typeMappings[reqType] = reqName
}

// ByName retrieves the reflect.Type associated with the given request name (reqName).
//
// Parameters:
//   - reqName: A string representing the name of the request.
//
// Returns:
//   - reqType: The reflect.Type associated with the given request name.
//   - err: An error if no mapping is cataloged for the given request name.
func (m *DefaultMappingCatalog) ByName(reqName string) (reqType reflect.Type, err error) {
	var ok bool
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if reqType, ok = m.nameMappings[reqName]; !ok {
		return nil, fmt.Errorf("%w for req name: %s", ErrMappingMissing, reqName)
	}
	return reqType, nil
}

// ByType retrieves the request name associated with the given request type (reqType).
//
// Parameters:
//   - reqType: A reflect.Type representing the type of the request.
//
// Returns:
//   - reqName: A string representing the name of the request associated with the given type.
//   - err: An error if no mapping is cataloged for the given request type.
func (m *DefaultMappingCatalog) ByType(reqType reflect.Type) (reqName string, err error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	var ok bool
	if reqName, ok = m.typeMappings[reqType]; !ok {
		return "", fmt.Errorf("%w for req type: %s", ErrMappingMissing, reqType)
	}
	return reqName, nil
}

// InsertMapping catalogs a mapping between a request name and its corresponding type.
//
// Type Parameters:
//   - TReq: The type of the command request, which must implement the CommandReq interface.
//
// Parameters:
//   - catalog: A pointer to the DefaultMappingCatalog where the mapping will be cataloged.
//   - reqName: A string representing the name of the request.
func InsertMapping[TReq CommandReq[CommandRes]](catalog *DefaultMappingCatalog, reqName string) {
	catalog.Insert(reqName, reflect.TypeFor[TReq]())
}
