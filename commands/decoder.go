package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dan-lugg/go-commands/util"
	"reflect"
	"sync"
)

var (
	ErrDecoderMissing = errors.New("decoder missing")
	ErrDecoderFailure = errors.New("decoder failure")
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
	decoders map[reflect.Type]Decoder
}

type NewDecoderCatalogOption = util.Option[*DecoderCatalog]

// NewDecoderCatalog creates and returns a new instance of DecoderCatalog.
// The catalog is initialized with an empty map for decoders, which associates
// reflect.Type with functions that decode serialized data into CommandReq[CommandRes].
func NewDecoderCatalog(options ...NewDecoderCatalogOption) (catalog *DecoderCatalog) {
	catalog = &DecoderCatalog{
		mutex:    sync.RWMutex{},
		decoders: make(map[reflect.Type]Decoder),
	}
	for _, option := range options {
		option(catalog)
	}
	return catalog
}

// Insert catalogs a decoder for a specific command request type.
//
// Parameters:
//   - reqName: The name of the request type to catalog.
//   - reqType: The reflect.Type of the request type.
//   - decoder: A Decoder function that decodes serialized data
//     into the specified command request type.
func (d *DecoderCatalog) Insert(reqType reflect.Type, decoder Decoder) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if d.decoders == nil {
		d.decoders = make(map[reflect.Type]Decoder)
	}
	d.decoders[reqType] = decoder
}

// InsertDecoder is a generic function that catalogs a decoder for a specific command request type.
//
// Parameters:
//   - catalog: A pointer to the DecoderCatalog where the decoder will be cataloged.
//   - reqName: The name of the request type to catalog.
//   - decoder: A Decoder function that decodes serialized data into the specified command request type.
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
func (d *DecoderCatalog) Decode(reqType reflect.Type, reqJSON []byte) (req CommandReq[CommandRes], err error) {
	d.mutex.RLock()
	decoder, found := d.decoders[reqType]
	d.mutex.RUnlock()
	if !found {
		return nil, fmt.Errorf("%w: req type: %s", ErrDecoderMissing, reqType)
	}
	req, err = decoder(reqJSON)
	if req == nil {
		return nil, fmt.Errorf("%w: req is nil", ErrDecoderFailure)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecoderFailure, err)
	}
	return req, nil
}
