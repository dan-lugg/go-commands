package commands

import (
	"context"
	"fmt"
	"github.com/dan-lugg/go-commands/util"
	"reflect"
	"sync"
)

// Handler is a generic interface for handling commands.
//
// Type Parameters:
//   - TReq: The type of the command request, which must implement the CommandReq interface.
//   - TRes: The type of the command response, which must implement the CommandRes interface.
//
// Methods:
//   - HandleRaw(req TReq, ctx context.Context) (res TRes, err error):
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
//   - HandleRaw(req CommandReq[CommandRes], ctx context.Context): Processes the given request (req) within the provided context (ctx),
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

// HandleRaw processes the given request (req) within the provided context (ctx).
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

// HandlerCatalog is a catalog for managing nameMappings between request types
// and their corresponding handler adapters.
//
// Fields:
//   - adapters: A map that associates reflect.Type with HandlerAdapter instances,
//     enabling the handling of specific request types.
type HandlerCatalog struct {
	mutex    sync.RWMutex
	adapters map[reflect.Type]HandlerAdapter
}

type NewHandlerCatalogOption = util.Option[*HandlerCatalog]

// NewHandlerCatalog creates and returns a new instance of HandlerCatalog.
//
// The catalog is initialized with an empty map for adapters, which associates
// reflect.Type with HandlerAdapter instances, enabling the handling of specific request types.
//
// Returns:
//   - A pointer to a HandlerCatalog instance.
func NewHandlerCatalog(options ...NewHandlerCatalogOption) *HandlerCatalog {
	catalog := &HandlerCatalog{
		mutex:    sync.RWMutex{},
		adapters: make(map[reflect.Type]HandlerAdapter),
	}
	for _, option := range options {
		option(catalog)
	}
	return catalog
}

// Insert adds a HandlerAdapter to the HandlerCatalog.
//
// Parameters:
//   - adapter: The HandlerAdapter instance to catalog.
//
// Behavior:
//   - Associates the request type (ReqType) of the adapter with the adapter itself
//     in the adapters map of the catalog.
//   - Enables the catalog to handle requests of the cataloged type using the adapter.
func (r *HandlerCatalog) Insert(adapter HandlerAdapter) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.adapters == nil {
		r.adapters = make(map[reflect.Type]HandlerAdapter)
	}
	r.adapters[adapter.ReqType()] = adapter
}

// HandleRaw processes a command request using the cataloged handler.
//
// Parameters:
//   - req: A CommandReq[CommandRes] representing the command request to be processed.
//   - ctx: A context.Context providing context for the request processing.
//
// Returns:
//   - res: A CommandRes representing the result of the command processing.
//   - err: An error if no handler is cataloged for the request type or if the handler fails.
//
// Behavior:
//   - Determines the type of the request (reqType) using reflection.
//   - Looks up the handler associated with the reqType in the catalog's adapters map.
//   - If no handler is found, returns an error indicating the request type is not cataloged.
//   - Delegates the request processing to the found handler and returns the result or an error.
func (r *HandlerCatalog) Handle(req CommandReq[CommandRes], ctx context.Context) (res CommandRes, err error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	reqType := reflect.TypeOf(req)
	handler, found := r.adapters[reqType]
	if !found {
		return nil, fmt.Errorf("no handler for reqType: %s, %w", reqType, util.ErrNotCataloged)
	}
	return handler.Handle(req, ctx)
}

// InsertHandler is a generic function that catalogs a handler for a specific command request type.
//
// Type Parameters:
//   - TReq: The type of the command request, which must implement the CommandReq interface.
//   - TRes: The type of the command response, which must implement the CommandRes interface.
//
// Parameters:
//   - catalog: A pointer to the HandlerCatalog where the handler will be cataloged.
//   - factory: A HandlerFactory function that creates a new instance of a Handler for the specified request and response types.
//
// Behavior:
//   - Creates a new DefaultHandlerAdapter using the provided factory function.
//   - Inserts the adapter in the given HandlerCatalog.
func InsertHandler[TReq CommandReq[TRes], TRes CommandRes](catalog *HandlerCatalog, factory HandlerFactory[TReq, TRes]) {
	catalog.Insert(NewDefaultHandlerAdapter(factory))
}

// TypeMap returns a mapping of request types to their corresponding response types.
//
// The method iterates over the cataloged adapters in the HandlerCatalog
// and constructs a map where the keys are the request types (reflect.Type)
// and the values are the response types (reflect.Type) produced by the adapters.
//
// Returns:
//   - typeMap: A map associating request types with their corresponding response types.
func (r *HandlerCatalog) TypeMap() (typeMap map[reflect.Type]reflect.Type) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	typeMap = make(map[reflect.Type]reflect.Type, len(r.adapters))
	for reqType, adapter := range r.adapters {
		typeMap[reqType] = adapter.ResType()
	}
	return typeMap
}
