package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
)

// <editor-fold desc="Command">

// CommandRes is an interface that represents the result of a command.
// It can be implemented by any type that represents the output of a command.
type CommandRes interface{}

// CommandReq is a generic interface representing a request for a command.
// It is parameterized by TRes, which must implement the CommandRes interface.
type CommandReq[TRes CommandRes] interface{}

// </editor-fold>

// <editor-fold desc="Decoder">

// CommandReqDecoder is a function type that takes a byte slice as input
// and returns a CommandReq[CommandRes] and an error. It is used to decode
// serialized command request data into a specific command request type.
type CommandReqDecoder func([]byte) (CommandReq[CommandRes], error)

// DefaultCommandReqDecoder returns a CommandReqDecoder function for decoding
// serialized command request data into a specific command request type.
// The generic type TReq must implement the CommandReq[CommandRes] interface.
//
// The returned decoder function takes a byte slice as input, attempts to
// unmarshal it into the specified TReq type, and returns the decoded
// command request or an error if unmarshalling fails.
func DefaultCommandReqDecoder[TReq CommandReq[CommandRes]]() CommandReqDecoder {
	return func(data []byte) (CommandReq[CommandRes], error) {
		var commandReq TReq
		if err := json.Unmarshal(data, &commandReq); err != nil {
			return nil, err
		}
		return commandReq, nil
	}
}

// DecoderRegistry is a registry for managing mappings between request names,
// their corresponding types, and decoders. It allows decoding serialized
// command request data into specific command request types.
//
// Fields:
//   - mappings: A map that associates request names (strings) with their
//     corresponding reflect.Type.
//   - decoders: A map that associates reflect.Type with functions that
//     decode serialized data into CommandReq[CommandRes].
type DecoderRegistry struct {
	mutex    sync.RWMutex
	mappings map[string]reflect.Type
	decoders map[reflect.Type]func([]byte) (CommandReq[CommandRes], error)
}

// NewDecoderRegistry creates and returns a new instance of DecoderRegistry.
// The registry is initialized with an empty map for decoders, which associates
// reflect.Type with functions that decode serialized data into CommandReq[CommandRes].
func NewDecoderRegistry() *DecoderRegistry {
	return &DecoderRegistry{
		mutex:    sync.RWMutex{},
		mappings: make(map[string]reflect.Type),
		decoders: make(map[reflect.Type]func([]byte) (CommandReq[CommandRes], error)),
	}
}

// RegisterDecoder registers a decoder for a specific command request type.
//
// Parameters:
//   - reqName: The name of the request type to register.
//   - reqType: The reflect.Type of the request type.
//   - decoder: A CommandReqDecoder function that decodes serialized data
//     into the specified command request type.
//
// Behavior:
//   - Initializes the mappings and decoders maps if they are nil.
//   - Associates the reqName with the reqType in the mappings map.
//   - Associates the reqType with the decoder function in the decoders map.
func (d *DecoderRegistry) RegisterDecoder(reqName string, reqType reflect.Type, decoder CommandReqDecoder) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if d.mappings == nil {
		d.mappings = make(map[string]reflect.Type)
	}
	if d.decoders == nil {
		d.decoders = make(map[reflect.Type]func([]byte) (CommandReq[CommandRes], error))
	}
	d.mappings[reqName] = reqType
	d.decoders[reqType] = decoder
}

// RegisterDecoder is a generic function that registers a decoder for a specific command request type.
//
// Parameters:
//   - registry: A pointer to the DecoderRegistry where the decoder will be registered.
//   - reqName: The name of the request type to register.
//   - decoder: A CommandReqDecoder function that decodes serialized data into the specified command request type.
//
// Behavior:
//   - Associates the reqName with the reflect.Type of the generic type TReq in the registry.
//   - Registers the provided decoder function for the TReq type in the registry.
func RegisterDecoder[TReq CommandReq[CommandRes]](registry *DecoderRegistry, reqName string, decoder CommandReqDecoder) {
	registry.RegisterDecoder(reqName, reflect.TypeFor[TReq](), decoder)
}

// Decode attempts to decode serialized command request data into a specific command request type.
//
// Parameters:
//   - reqName: The name of the request type to decode.
//   - reqData: A byte slice containing the serialized command request data.
//
// Returns:
//   - A CommandReq[CommandRes] representing the decoded command request.
//   - An error if the decoding fails or if no decoder is registered for the given request name.
//
// Behavior:
//   - Looks up the reqType associated with the reqName in the mappings map.
//   - If no reqType is found, returns an error indicating the request name is not registered.
//   - Retrieves the decoder function associated with the reqType from the decoders map.
//   - If no decoder is found, returns an error indicating the type is not registered.
//   - Uses the decoder function to decode the reqData into the corresponding command request type.
func (d *DecoderRegistry) Decode(reqName string, reqData []byte) (CommandReq[CommandRes], error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	reqType, found := d.mappings[reqName]
	if !found {
		return nil, fmt.Errorf("no command type registered for reqName: %s", reqName)
	}
	factory, found := d.decoders[reqType]
	if !found {
		return nil, fmt.Errorf("no command decoder registered for type: %s", reqData)
	}
	return factory(reqData)
}

// Resolver is a generic interface for resolving input data into a request name and request data.
//
// Type Parameters:
//   - TAny: The type of the input data to be resolved.
//
// Methods:
//   - Resolve(input TAny) (reqName string, reqData []byte, err error):
//     Resolves the input data into a request name (reqName) and request data (reqData).
//     Returns an error (err) if the resolution fails.
type Resolver[TAny any] interface {
	Resolve(input TAny) (reqName string, reqData []byte, err error)
}

// JSONResolver is a type that implements the Resolver interface for resolving
// JSON input data into a request name and request data.
//
// Fields:
//   - Resolver[[]byte]: Embeds the generic Resolver interface for handling
//     byte slice input.
//
// Methods:
//   - Resolve(input []byte) (reqName string, reqData []byte, err error):
//     Resolves the input JSON data into a request name (reqName) and request
//     data (reqData). Returns an error (err) if the resolution fails.
type JSONResolver struct {
	Resolver[[]byte]
}

// Resolve parses the input JSON data and extracts the request name (reqName) and request data (reqData).
//
// Parameters:
//   - input: A byte slice containing the JSON data to be resolved.
//
// Returns:
//   - reqName: The name of the request type extracted from the JSON.
//   - reqData: The raw JSON data associated with the request.
//   - err: An error if the JSON unmarshalling fails.
//
// Behavior:
//   - Unmarshals the input JSON into a struct containing TypeName and reqData fields.
//   - Returns the TypeName as reqName and the reqData as reqData.
//   - If unmarshalling fails, returns an error indicating the failure.
func (b *JSONResolver) Resolve(input []byte) (reqName string, reqData []byte, err error) {
	var wrapped struct {
		TypeName string          `json:"type"`
		Data     json.RawMessage `json:"reqData"`
	}
	if err = json.Unmarshal(input, &wrapped); err != nil {
		return "", nil, fmt.Errorf("failed resolve reqName and reqData: %w", err)
	}
	return wrapped.TypeName, wrapped.Data, nil
}

// </editor-fold>

// <editor-fold desc="Handler">

// Handler is a generic interface for handling commands.
//
// Type Parameters:
//   - TReq: The type of the command request, which must implement the CommandReq interface.
//   - TRes: The type of the command response, which must implement the CommandRes interface.
//
// Methods:
//   - Handle(req TReq, ctx context.Context) (res TRes, err error):
//     Processes the given command request (req) within the provided context (ctx).
//     Returns the command response (res) and an error (err) if the handling fails.
type Handler[TReq CommandReq[TRes], TRes CommandRes] interface {
	Handle(req TReq, ctx context.Context) (res TRes, err error)
}

// HandlerFactory is a type alias for a function that creates a new instance of a Handler.
//
// Type Parameters:
//   - TReq: The type of the command request, which must implement the CommandReq interface.
//   - TRes: The type of the command response, which must implement the CommandRes interface.
//
// Returns:
//   - A Handler instance capable of processing the specified request and producing the corresponding response.
type HandlerFactory[TReq CommandReq[TRes], TRes CommandRes] func() Handler[TReq, TRes]

// HandlerAdapter is an interface for adapting handlers to a common structure.
//
// Methods:
//   - ReqType(): Returns the reflect.Type of the request handled by the adapter.
//   - ResType(): Returns the reflect.Type of the response produced by the adapter.
//   - Handle(req CommandReq[CommandRes], ctx context.Context): Processes the given request (req) within the provided context (ctx),
//     returning the response (res) or an error (err) if the handling fails.
type HandlerAdapter interface {
	ReqType() reflect.Type
	ResType() reflect.Type
	Handle(req CommandReq[CommandRes], ctx context.Context) (res CommandRes, err error)
}

// DefaultHandlerAdapter is a generic adapter for handling commands.
//
// Type Parameters:
//   - TReq: The type of the command request, which must implement the CommandReq interface.
//   - TRes: The type of the command response, which must implement the CommandRes interface.
//
// Fields:
//   - handler: An instance of the Handler that processes the command request.
//   - handlerFactory: A factory function that creates a new instance of the Handler.
type DefaultHandlerAdapter[TReq CommandReq[TRes], TRes CommandRes] struct {
	mutex          sync.RWMutex
	handler        Handler[TReq, TRes]
	handlerFactory HandlerFactory[TReq, TRes]
}

// NewDefaultHandlerAdapter creates a new instance of DefaultHandlerAdapter.
//
// Type Parameters:
//   - TReq: The type of the command request, which must implement the CommandReq interface.
//   - TRes: The type of the command response, which must implement the CommandRes interface.
//
// Parameters:
//   - factory: A function that creates a new instance of a Handler for the specified request and response types.
//
// Returns:
//   - A pointer to a DefaultHandlerAdapter instance, initialized with the provided factory function.
func NewDefaultHandlerAdapter[TReq CommandReq[TRes], TRes CommandRes](factory func() Handler[TReq, TRes]) *DefaultHandlerAdapter[TReq, TRes] {
	return &DefaultHandlerAdapter[TReq, TRes]{
		mutex:          sync.RWMutex{},
		handler:        nil,
		handlerFactory: factory,
	}
}

// Handle processes the given request (req) within the provided context (ctx).
//
// Type Parameters:
//   - TReq: The type of the command request, which must implement the CommandReq interface.
//   - TRes: The type of the command response, which must implement the CommandRes interface.
//
// Parameters:
//   - req: A CommandReq[CommandRes] representing the command request to be processed.
//   - ctx: A context.Context providing context for the request processing.
//
// Returns:
//   - res: A CommandRes representing the result of the command processing.
//   - err: An error if the request type does not match the expected type or if the handler fails.
//
// Behavior:
//   - Attempts to cast the req to the expected type TReq.
//   - If the cast fails, returns an error indicating the type mismatch.
//   - Lazily initializes the handler using the handlerFactory if it is not already initialized.
//   - Delegates the request processing to the handler and returns the result or an error.
func (a *DefaultHandlerAdapter[TReq, TRes]) Handle(req CommandReq[CommandRes], ctx context.Context) (res CommandRes, err error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	adaptedReq, ok := req.(TReq)
	if !ok {
		return nil, fmt.Errorf("request type %T does not match expected type %T", req, a.handler)
	}
	if a.handler == nil {
		a.handler = a.handlerFactory()
	}
	return a.handler.Handle(adaptedReq, ctx)
}

// ReqType returns the reflect.Type of the request handled by the adapter.
//
// Type Parameters:
//   - TReq: The type of the command request, which must implement the CommandReq interface.
//
// Returns:
//   - A reflect.Type representing the type of the request handled by the adapter.
func (a *DefaultHandlerAdapter[TReq, TRes]) ReqType() reflect.Type {
	return reflect.TypeFor[TReq]()
}

// ResType returns the reflect.Type of the response produced by the adapter.
//
// Type Parameters:
//   - TRes: The type of the command response, which must implement the CommandRes interface.
//
// Returns:
//   - A reflect.Type representing the type of the response produced by the adapter.
func (a *DefaultHandlerAdapter[TReq, TRes]) ResType() reflect.Type {
	return reflect.TypeFor[TRes]()
}

// HandlerRegistry is a registry for managing mappings between request types
// and their corresponding handler adapters.
//
// Fields:
//   - adapters: A map that associates reflect.Type with HandlerAdapter instances,
//     enabling the handling of specific request types.
type HandlerRegistry struct {
	mutex    sync.RWMutex
	adapters map[reflect.Type]HandlerAdapter
}

// NewHandlerRegistry creates and returns a new instance of HandlerRegistry.
//
// The registry is initialized with an empty map for adapters, which associates
// reflect.Type with HandlerAdapter instances, enabling the handling of specific request types.
//
// Returns:
//   - A pointer to a HandlerRegistry instance.
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		mutex:    sync.RWMutex{},
		adapters: make(map[reflect.Type]HandlerAdapter),
	}
}

// Register adds a HandlerAdapter to the HandlerRegistry.
//
// Parameters:
//   - adapter: The HandlerAdapter instance to register.
//
// Behavior:
//   - Associates the request type (ReqType) of the adapter with the adapter itself
//     in the adapters map of the registry.
//   - Enables the registry to handle requests of the registered type using the adapter.
func (r *HandlerRegistry) Register(adapter HandlerAdapter) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.adapters[adapter.ReqType()] = adapter
}

// Handle processes a command request using the registered handler.
//
// Parameters:
//   - req: A CommandReq[CommandRes] representing the command request to be processed.
//   - ctx: A context.Context providing context for the request processing.
//
// Returns:
//   - res: A CommandRes representing the result of the command processing.
//   - err: An error if no handler is registered for the request type or if the handler fails.
//
// Behavior:
//   - Determines the type of the request (reqType) using reflection.
//   - Looks up the handler associated with the reqType in the registry's adapters map.
//   - If no handler is found, returns an error indicating the request type is not registered.
//   - Delegates the request processing to the found handler and returns the result or an error.
func (r *HandlerRegistry) Handle(req CommandReq[CommandRes], ctx context.Context) (res CommandRes, err error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	reqType := reflect.TypeOf(req)
	handler, found := r.adapters[reqType]
	if !found {
		return nil, fmt.Errorf("no handler registered for request type: %s", reqType)
	}
	return handler.Handle(req, ctx)
}

// RegisterHandler is a generic function that registers a handler for a specific command request type.
//
// Type Parameters:
//   - TReq: The type of the command request, which must implement the CommandReq interface.
//   - TRes: The type of the command response, which must implement the CommandRes interface.
//
// Parameters:
//   - registry: A pointer to the HandlerRegistry where the handler will be registered.
//   - factory: A HandlerFactory function that creates a new instance of a Handler for the specified request and response types.
//
// Behavior:
//   - Creates a new DefaultHandlerAdapter using the provided factory function.
//   - Registers the adapter in the given HandlerRegistry.
func RegisterHandler[TReq CommandReq[TRes], TRes CommandRes](registry *HandlerRegistry, factory HandlerFactory[TReq, TRes]) {
	registry.Register(NewDefaultHandlerAdapter(factory))
}

// HandlerInfo represents metadata about a handler, including its request name,
// request type, and response type.
//
// Fields:
//   - ReqName: The name of the request type handled by the handler.
//   - ReqType: The reflect.Type of the request type.
//   - ResType: The reflect.Type of the response type.
type HandlerInfo struct {
	ReqName string
	ReqType reflect.Type
	ResType reflect.Type
}

// Infos retrieves metadata about all registered handlers in the HandlerRegistry.
//
// Returns:
//   - infos: A slice of HandlerInfo containing details about each registered handler,
//     including its request name, request type, and response type.
//
// Behavior:
//   - Iterates over all registered handler adapters in the registry.
//   - Constructs a HandlerInfo for each adapter, including its request name,
//     request type, and response type.
//   - Returns the slice of HandlerInfo containing metadata for all handlers.
//
// Note:
//   - The ReqName field currently uses the name of the request type (ReqType.Name()),
//     but it may need to be mapped from the decoders for more accurate naming.
func (r *HandlerRegistry) Infos() (infos []HandlerInfo) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	infos = make([]HandlerInfo, 0, len(r.adapters))
	for _, adapter := range r.adapters {
		info := HandlerInfo{
			ReqName: adapter.ReqType().Name(), // TODO: Need to map this from the decoders somehow
			ReqType: adapter.ReqType(),
			ResType: adapter.ResType(),
		}
		infos = append(infos, info)
	}
	return infos
}

// </editor-fold>
