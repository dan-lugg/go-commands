package openapi

import (
	"encoding/json"
	"fmt"
	"github.com/dan-lugg/go-commands/commands"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"io"
	"reflect"
)

type SpecWriter struct {
	mappingRegistry *commands.MappingRegistry
	handlerRegistry *commands.HandlerRegistry
}

func NewSpecWriter(mappingRegistry *commands.MappingRegistry, handlerRegistry *commands.HandlerRegistry) *SpecWriter {
	return &SpecWriter{
		mappingRegistry: mappingRegistry,
		handlerRegistry: handlerRegistry,
	}
}

func (w *SpecWriter) WriteSpec(writer io.Writer) (err error) {
	paths := openapi3.NewPaths()

	for reqType, resType := range w.handlerRegistry.TypeMap() {
		var reqName string
		var pathItem openapi3.PathItem

		reqName, err = w.mappingRegistry.ByType(reqType)
		if err != nil {
			return fmt.Errorf("failed to get request name for type %s: %w", reqType.Name(), err)
		}

		pathItem, err = w.CreatePathItem(reqName, reqType, resType)
		if err != nil {
			return fmt.Errorf("failed to create path item for request type %s: %w", reqType.Name(), err)
		}

		paths.Set(fmt.Sprintf("/%s", reqName), &pathItem)
	}

	encoder := json.NewEncoder(writer)
	err = encoder.Encode(paths)
	if err != nil {
		return fmt.Errorf("failed to encode OpenAPI paths: %w", err)
	}
	return nil
}

func (w *SpecWriter) CreatePathItem(reqName string, reqType reflect.Type, resType reflect.Type) (pathItem openapi3.PathItem, err error) {
	var reqSchemaRef *openapi3.SchemaRef
	var resSchemaRef *openapi3.SchemaRef

	generator := openapi3gen.NewGenerator()

	reqSchemaRef, err = generator.GenerateSchemaRef(reqType)
	if err != nil {
		return openapi3.PathItem{}, fmt.Errorf("failed to generate schema for request type %s: %w", reqType.Name(), err)
	}

	resSchemaRef, err = generator.GenerateSchemaRef(resType)
	if err != nil {
		return openapi3.PathItem{}, fmt.Errorf("failed to generate schema for response type %s: %w", resType.Name(), err)
	}

	operation := &openapi3.Operation{
		Summary:     fmt.Sprintf("Handle %s", reqName),
		Description: fmt.Sprintf("Handles the %s command", reqName),
		OperationID: reqName,
	}

	operation.RequestBody = &openapi3.RequestBodyRef{
		Value: &openapi3.RequestBody{
			Required: true,
			Content:  openapi3.NewContentWithJSONSchema(reqSchemaRef.Value),
		},
	}
	operation.AddResponse(200, openapi3.NewResponse().
		WithContent(openapi3.NewContentWithJSONSchema(resSchemaRef.Value)))

	return openapi3.PathItem{
		Post: operation,
	}, nil
}
