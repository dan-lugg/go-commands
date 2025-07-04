package commands

import (
	"encoding/json"
	"fmt"
	"github.com/dan-lugg/go-commands/util"
	"reflect"
	"sync"
)

// Decoder is a function type that takes a byte slice as input
// and returns a CommandReq[CommandRes] and an error. It is used to decode
// serialized command request data into a specific command request type.
type Decoder func([]byte) (CommandReq[CommandRes], error)

// DefaultDecoder returns a Decoder function for decoding
// serialized command request data into a specific command request type.
// The generic type TReq must implement the CommandReq[CommandRes] interface.
//
// The returned decoder function takes a byte slice as input, attempts to
// unmarshal it into the specified TReq type, and returns the decoded
// command request or an error if unmarshalling fails.
func DefaultDecoder[TReq CommandReq[CommandRes]]() Decoder {
	return func(data []byte) (CommandReq[CommandRes], error) {
		var commandReq TReq
		if err := json.Unmarshal(data, &commandReq); err != nil {
			return nil, err
		}
		return commandReq, nil
	}
}

// DecoderCatalog is a catalog for managing nameMappings between request names,
// their corresponding types, and decoders. It allows decoding serialized
// command request data into specific command request types.
//
// Fields:
//   - nameMappings: A map that associates request names (strings) with their
//     corresponding reflect.Type.
//   - decoders: A map that associates reflect.Type with functions that
//     decode serialized data into CommandReq[CommandRes].
type DecoderCatalog struct {
	mutex    sync.RWMutex
	decoders map[reflect.Type]func([]byte) (CommandReq[CommandRes], error)
}

type NewDecoderCatalogOption = util.Option[*DecoderCatalog]

// NewDecoderCatalog creates and returns a new instance of DecoderCatalog.
// The catalog is initialized with an empty map for decoders, which associates
// reflect.Type with functions that decode serialized data into CommandReq[CommandRes].
func NewDecoderCatalog(options ...NewDecoderCatalogOption) (catalog *DecoderCatalog) {
	catalog = &DecoderCatalog{
		mutex:    sync.RWMutex{},
		decoders: make(map[reflect.Type]func([]byte) (CommandReq[CommandRes], error)),
	}
	for _, option := range options {
		option(catalog)
	}
	return catalog
}

// InsertDecoder catalogs a decoder for a specific command request type.
//
// Parameters:
//   - reqName: The name of the request type to catalog.
//   - reqType: The reflect.Type of the request type.
//   - decoder: A Decoder function that decodes serialized data
//     into the specified command request type.
//
// Behavior:
//   - Initializes the nameMappings and decoders maps if they are nil.
//   - Associates the reqName with the reqType in the nameMappings map.
//   - Associates the reqType with the decoder function in the decoders map.
func (d *DecoderCatalog) Insert(reqType reflect.Type, decoder Decoder) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if d.decoders == nil {
		d.decoders = make(map[reflect.Type]func([]byte) (CommandReq[CommandRes], error))
	}
	d.decoders[reqType] = decoder
}

// InsertDecoder is a generic function that catalogs a decoder for a specific command request type.
//
// Parameters:
//   - catalog: A pointer to the DecoderCatalog where the decoder will be cataloged.
//   - reqName: The name of the request type to catalog.
//   - decoder: A Decoder function that decodes serialized data into the specified command request type.
//
// Behavior:
//   - Associates the reqName with the reflect.Type of the generic type TReq in the catalog.
//   - Inserts the provided decoder function for the TReq type in the catalog.
func InsertDecoder[TReq CommandReq[CommandRes]](catalog *DecoderCatalog, decoder Decoder) {
	catalog.Insert(reflect.TypeFor[TReq](), decoder)
}

// Decode attempts to decode serialized command request data into a specific command request type.
//
// Parameters:
//   - reqName: The name of the request type to decode.
//   - reqJSON: A byte slice containing the serialized command request data.
//
// Returns:
//   - A CommandReq[CommandRes] representing the decoded command request.
//   - An error if the decoding fails or if no decoder is cataloged for the given request name.
//
// Behavior:
//   - Looks up the reqType associated with the reqName in the nameMappings map.
//   - If no reqType is found, returns an error indicating the request name is not cataloged.
//   - Retrieves the decoder function associated with the reqType from the decoders map.
//   - If no decoder is found, returns an error indicating the type is not cataloged.
//   - Uses the decoder function to decode the reqJSON into the corresponding command request type.
func (d *DecoderCatalog) Decode(reqType reflect.Type, reqJSON []byte) (CommandReq[CommandRes], error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	factory, found := d.decoders[reqType]
	if !found {
		return nil, fmt.Errorf("no decoder for reqType: %s, %w", reqType, util.ErrNotCataloged)
	}
	return factory(reqJSON)
}
