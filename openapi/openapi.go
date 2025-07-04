package openapi

import (
	"fmt"
	"github.com/dan-lugg/go-commands/commands"
	"github.com/dan-lugg/go-commands/util"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"io"
	"reflect"
)

type SpecWriter struct {
	title          string
	version        string
	description    string
	mappingCatalog *commands.MappingCatalog
	handlerCatalog *commands.HandlerCatalog
}

type SpecWriterOption = util.Option[*SpecWriter]

func WithTitle(title string) SpecWriterOption {
	return func(w *SpecWriter) {
		w.title = title
	}
}

func WithVersion(version string) SpecWriterOption {
	return func(w *SpecWriter) {
		w.version = version
	}
}

func WithDescription(description string) SpecWriterOption {
	return func(w *SpecWriter) {
		w.description = description
	}
}

func NewSpecWriter(mappingCatalog *commands.MappingCatalog, handlerCatalog *commands.HandlerCatalog, options ...SpecWriterOption) (specWriter *SpecWriter) {
	specWriter = &SpecWriter{
		title:          "Commands API",
		version:        "1.0.0",
		description:    "API for handling commands",
		mappingCatalog: mappingCatalog,
		handlerCatalog: handlerCatalog,
	}
	for _, option := range options {
		option(specWriter)
	}
	return specWriter
}

func (w *SpecWriter) WriteSpec(writer io.Writer) (err error) {
	spec, err := w.CreateSpec()
	if err != nil {
		return fmt.Errorf("failed to create OpenAPI spec: %w", err)
	}

	data, err := spec.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal OpenAPI spec to JSON: %w", err)
	}

	size, err := writer.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write OpenAPI spec to writer: %w", err)
	}
	_ = size

	return nil
}

func (w *SpecWriter) CreateSpec() (spec openapi3.T, err error) {
	spec = openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:       w.title,
			Version:     w.version,
			Description: w.description,
		},
	}

	spec.Paths = openapi3.NewPaths()

	for reqType, resType := range w.handlerCatalog.TypeMap() {
		var reqName string
		var pathItem openapi3.PathItem

		reqName, err = w.mappingCatalog.ByType(reqType)
		if err != nil {
			return openapi3.T{}, fmt.Errorf("failed to get request name for type %s: %w", reqType.Name(), err)
		}

		pathItem, err = w.CreatePathItem(reqName, reqType, resType)
		if err != nil {
			return openapi3.T{}, fmt.Errorf("failed to create path item for request type %s: %w", reqType.Name(), err)
		}

		spec.Paths.Set(fmt.Sprintf("/%s", reqName), &pathItem)
	}

	return spec, nil
}

func (w *SpecWriter) CreatePathItem(reqName string, reqType reflect.Type, resType reflect.Type) (pathItem openapi3.PathItem, err error) {
	generator := openapi3gen.NewGenerator(
		openapi3gen.CreateComponentSchemas(openapi3gen.ExportComponentSchemasOptions{
			ExportComponentSchemas: false,
			ExportTopLevelSchema:   false,
			ExportGenerics:         false,
		}),
	)

	var reqSchemaRef *openapi3.SchemaRef
	var resSchemaRef *openapi3.SchemaRef

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
