package commands

import (
	"context"
	"fmt"
	"github.com/dan-lugg/go-commands/util"
)

type NewManagerOption = util.Option[*Manager]

type Manager struct {
	mappingCatalog *MappingCatalog
	decoderCatalog *DecoderCatalog
	handlerCatalog *HandlerCatalog
}

func NewManager(mappingCatalog *MappingCatalog, decoderCatalog *DecoderCatalog, handlerCatalog *HandlerCatalog, options ...NewManagerOption) *Manager {
	manager := &Manager{
		mappingCatalog: mappingCatalog,
		decoderCatalog: decoderCatalog,
		handlerCatalog: handlerCatalog,
	}
	for _, option := range options {
		option(manager)
	}
	return manager
}

func Insert[TReq CommandReq[TRes], TRes CommandRes](manager *Manager, reqName string, decoder Decoder, factory HandlerFactory[TReq, TRes]) {
	InsertMapping[TReq](manager.mappingCatalog, reqName)
	InsertDecoder[TReq](manager.decoderCatalog, decoder)
	InsertHandler[TReq, TRes](manager.handlerCatalog, factory)
}

func (manager *Manager) HandleRaw(reqName string, reqJSON []byte, ctx context.Context) (res CommandRes, err error) {
	reqType, err := manager.mappingCatalog.ByName(reqName)
	if err != nil {
		return nil, fmt.Errorf("error mapping request type by name: %w", err)
	}
	req, err := manager.decoderCatalog.Decode(reqType, reqJSON)
	if err != nil {
		return nil, fmt.Errorf("error decoding request: %w", err)
	}
	res, err = manager.handlerCatalog.Handle(req, ctx)
	if err != nil {
		return nil, fmt.Errorf("error handling request: %w", err)
	}
	return res, nil
}

func (manager *Manager) HandleReq(req CommandReq[CommandRes], ctx context.Context) (res CommandRes, err error) {
	res, err = manager.handlerCatalog.Handle(req, ctx)
	if err != nil {
		return nil, fmt.Errorf("error handling request: %w", err)
	}
	return res, nil
}

func HandleRaw(manager *Manager, reqName string, reqJSON []byte, ctx context.Context) (res CommandRes, err error) {
	res, err = manager.HandleRaw(reqName, reqJSON, ctx)
	if err != nil {
		return res, fmt.Errorf("error handling request: %w", err)
	}
	return res, nil
}

func HandleReq[TReq CommandReq[TRes], TRes CommandRes](manager *Manager, req TReq, ctx context.Context) (res TRes, err error) {
	var genericRes CommandRes
	genericRes, err = manager.HandleReq(req, ctx)
	if err != nil {
		return res, fmt.Errorf("error handling request: %w", err)
	}
	res = genericRes.(TRes)
	return res, nil
}
