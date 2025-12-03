package commands

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/dan-lugg/go-commands/pkg/futures"
	"github.com/dan-lugg/go-commands/pkg/util"
)

var (
	ErrDecoderMissing = errors.New("decoder missing")
	ErrDecoderFailure = errors.New("decoder failure")
	ErrHandlerMissing = errors.New("handler missing")
	ErrInvalidReqType = errors.New("invalid req type")
	ErrInvalidResType = errors.New("invalid res type")
)

type CommandRes any

type CommandReq[TRes CommandRes] any

type Handler[TReq CommandReq[TRes], TRes CommandRes] interface {
	Handle(ctx context.Context, req TReq) (res TRes, err error)
}

type HandlerFactoryFunc[TReq CommandReq[TRes], TRes CommandRes] func() Handler[TReq, TRes]

type HandlerAdapter interface {
	ReqType() reflect.Type
	ResType() reflect.Type
	Handle(ctx context.Context, req CommandReq[CommandRes]) (res CommandRes, err error)
}

type DefaultHandlerAdapter[TReq CommandReq[TRes], TRes CommandRes] struct {
	mutex          sync.RWMutex
	handler        Handler[TReq, TRes]
	handlerFactory HandlerFactoryFunc[TReq, TRes]
}

func NewDefaultHandlerAdapter[TReq CommandReq[TRes], TRes CommandRes](factory HandlerFactoryFunc[TReq, TRes]) *DefaultHandlerAdapter[TReq, TRes] {
	return &DefaultHandlerAdapter[TReq, TRes]{
		mutex:          sync.RWMutex{},
		handler:        nil,
		handlerFactory: factory,
	}
}

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

func (a *DefaultHandlerAdapter[TReq, TRes]) ReqType() reflect.Type {
	return reflect.TypeFor[TReq]()
}

func (a *DefaultHandlerAdapter[TReq, TRes]) ResType() reflect.Type {
	return reflect.TypeFor[TRes]()
}

type HandlerCatalog interface {
	Insert(adapter HandlerAdapter)
	Handle(ctx context.Context, req CommandReq[CommandRes]) (res CommandRes, err error)
	Future(ctx context.Context, req CommandReq[CommandRes]) futures.Future[util.Tuple2[CommandRes, error]]
	TypeMap() map[reflect.Type]reflect.Type
}

type DefaultHandlerCatalog struct {
	mutex    sync.RWMutex
	adapters map[reflect.Type]HandlerAdapter
}

type NewDefaultHandlerCatalogOption = util.Option[*DefaultHandlerCatalog]

func NewDefaultHandlerCatalog(options ...NewDefaultHandlerCatalogOption) *DefaultHandlerCatalog {
	catalog := &DefaultHandlerCatalog{
		mutex:    sync.RWMutex{},
		adapters: make(map[reflect.Type]HandlerAdapter),
	}
	for _, option := range options {
		option(catalog)
	}
	return catalog
}

func (r *DefaultHandlerCatalog) Insert(adapter HandlerAdapter) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.adapters == nil {
		r.adapters = make(map[reflect.Type]HandlerAdapter)
	}
	r.adapters[adapter.ReqType()] = adapter
}

func (r *DefaultHandlerCatalog) Handle(ctx context.Context, req CommandReq[CommandRes]) (res CommandRes, err error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	reqType := reflect.TypeOf(req)
	adapter, found := r.adapters[reqType]
	if !found {
		return nil, fmt.Errorf("%w for req type: %s", ErrHandlerMissing, reqType)
	}
	return adapter.Handle(ctx, req)
}

func Handle[TReq CommandReq[TRes], TRes CommandRes](ctx context.Context, catalog *DefaultHandlerCatalog, req TReq) (typedRes TRes, err error) {
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

func (r *DefaultHandlerCatalog) Future(ctx context.Context, req CommandReq[CommandRes]) futures.Future[util.Tuple2[CommandRes, error]] {
	return futures.Start(ctx, func(ctx context.Context) util.Tuple2[CommandRes, error] {
		res, err := r.Handle(ctx, req)
		return util.Tuple2[CommandRes, error]{
			Val1: res,
			Val2: err,
		}
	})
}

func Future[TReq CommandReq[TRes], TRes CommandRes](ctx context.Context, catalog *DefaultHandlerCatalog, req TReq) futures.Future[util.Tuple2[TRes, error]] {
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

func InsertHandler[TReq CommandReq[TRes], TRes CommandRes](catalog *DefaultHandlerCatalog, factory HandlerFactoryFunc[TReq, TRes]) {
	catalog.Insert(NewDefaultHandlerAdapter(factory))
}

func (r *DefaultHandlerCatalog) TypeMap() (typeMap map[reflect.Type]reflect.Type) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	typeMap = make(map[reflect.Type]reflect.Type, len(r.adapters))
	for reqType, adapter := range r.adapters {
		typeMap[reqType] = adapter.ResType()
	}
	return typeMap
}
