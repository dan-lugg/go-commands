package commands

import (
	"context"
	"errors"
	"fmt"
	"github.com/dan-lugg/go-commands/futures"
	"github.com/dan-lugg/go-commands/util"
	"reflect"
	"sync"
)

var (
	ErrHandlerMissing = errors.New("handler missing")
	ErrInvalidReqType = errors.New("invalid req type")
	ErrInvalidResType = errors.New("invalid res type")
)

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
	Handle(ctx context.Context, req TReq) (res TRes, err error)
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
	Handle(ctx context.Context, req CommandReq[CommandRes]) (res CommandRes, err error)
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
//   - ctx: A context.Context providing context for the request processing.
//   - req: A CommandReq[CommandRes] representing the command request to be processed.
//
// Returns:
//   - res: A CommandRes representing the result of the command processing.
//   - err: An error if the request type does not match the expected type or if the handler fails.
func (a *DefaultHandlerAdapter[TReq, TRes]) Handle(ctx context.Context, req CommandReq[CommandRes]) (res CommandRes, err error) {
	typedReq, ok := req.(TReq)
	if !ok {
		return nil, fmt.Errorf("req type %T does not match %T", req, typedReq)
	}
	a.mutex.RLock()
	handler := a.handler
	a.mutex.RUnlock()
	if handler == nil {
		func() {
			a.mutex.Lock()
			defer a.mutex.Unlock()
			if a.handler == nil {
				a.handler = a.handlerFactory()
			}
		}()
		handler = a.handler
	}
	if handler == nil {
		return nil, fmt.Errorf("%w for req type: %s", ErrHandlerMissing, a.ReqType())
	}
	return handler.Handle(ctx, typedReq)
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
func (r *HandlerCatalog) Insert(adapter HandlerAdapter) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.adapters == nil {
		r.adapters = make(map[reflect.Type]HandlerAdapter)
	}
	r.adapters[adapter.ReqType()] = adapter
}

// Handle processes a command request using the cataloged handler.
//
// Parameters:
//   - req: A CommandReq[CommandRes] representing the command request to be processed.
//   - ctx: A context.Context providing context for the request processing.
//
// Returns:
//   - res: A CommandRes representing the result of the command processing.
//   - err: An error if no handler is cataloged for the request type or if the handler fails.
func (r *HandlerCatalog) Handle(ctx context.Context, req CommandReq[CommandRes]) (res CommandRes, err error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	reqType := reflect.TypeOf(req)
	adapter, found := r.adapters[reqType]
	if !found {
		return nil, fmt.Errorf("%w for req type: %s", ErrHandlerMissing, reqType)
	}
	return adapter.Handle(ctx, req)
}

// Handle processes a command request using the cataloged handler.
//
// Type Parameters:
//   - TReq: The type of the command request, which must implement the CommandReq interface.
//   - TRes: The type of the command response, which must implement the CommandRes interface.
//
// Parameters:
//   - ctx: A context.Context providing context for the request processing.
//   - catalog: A pointer to the HandlerCatalog containing the cataloged handlers.
//   - req: A TReq representing the command request to be processed.
//
// Returns:
//   - res: A TRes representing the result of the command processing.
//   - err: An error if the request type does not match the expected type or if the handler fails.
func Handle[TReq CommandReq[TRes], TRes CommandRes](ctx context.Context, catalog *HandlerCatalog, req TReq) (typedRes TRes, err error) {
	res, err := catalog.Handle(ctx, req)
	if errors.Is(err, ErrHandlerMissing) {
		return *new(TRes), err
	}
	var ok bool
	if typedRes, ok = res.(TRes); !ok {
		return *new(TRes), fmt.Errorf("%w %T was unexpected for %T", ErrInvalidResType, res, typedRes)
	}
	return typedRes, err
}

// Future creates a futures.Future that asynchronously processes a command request.
//
// Parameters:
//   - ctx: A context.Context providing context for the request processing.
//   - req: A CommandReq[CommandRes] representing the command request to be processed.
//
// Returns:
//   - A futures.Future containing a util.Tuple2 where:
//   - Val1 is the CommandRes representing the result of the command processing.
//   - Val2 is an error if the processing fails.
func (r *HandlerCatalog) Future(ctx context.Context, req CommandReq[CommandRes]) futures.Future[util.Tuple2[CommandRes, error]] {
	return futures.Start(ctx, func(ctx context.Context) util.Tuple2[CommandRes, error] {
		res, err := r.Handle(ctx, req)
		return util.Tuple2[CommandRes, error]{
			Val1: res,
			Val2: err,
		}
	})
}

// Future creates a futures.Future that asynchronously processes a command request.
//
// Type Parameters:
//   - TReq: The type of the command request, which must implement the CommandReq interface.
//   - TRes: The type of the command response, which must implement the CommandRes interface.
//
// Parameters:
//   - ctx: A context.Context providing context for the request processing.
//   - catalog: A pointer to the HandlerCatalog containing the cataloged handlers.
//   - req: A TReq representing the command request to be processed.
//
// Returns:
//   - A futures.Future containing a util.Tuple2 where:
//   - Val1 is the TRes representing the result of the command processing.
//   - Val2 is an error if the processing fails.
func Future[TReq CommandReq[TRes], TRes CommandRes](ctx context.Context, catalog *HandlerCatalog, req TReq) futures.Future[util.Tuple2[TRes, error]] {
	return futures.Start(ctx, func(ctx context.Context) util.Tuple2[TRes, error] {
		tup := catalog.Future(ctx, req).Wait()
		res, err := tup.Val1, tup.Val2
		if err != nil {
			return util.Tuple2[TRes, error]{
				Val1: *new(TRes),
				Val2: err,
			}
		}
		typedRes, ok := res.(TRes)
		if !ok {
			return util.Tuple2[TRes, error]{
				Val1: *new(TRes),
				Val2: fmt.Errorf("%w %T was unexpected for %T", ErrInvalidResType, res, typedRes),
			}
		}
		return util.Tuple2[TRes, error]{
			Val1: typedRes,
			Val2: err,
		}
	})
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
