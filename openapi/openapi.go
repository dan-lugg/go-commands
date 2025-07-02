package openapi

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"go-commands-v2/commands"
)

func CreateOpenAPI3Paths(infos []commands.HandlerInfo) (paths *openapi3.Paths, err error) {
	generator := openapi3gen.NewGenerator()
	paths = openapi3.NewPaths()
	for _, info := range infos {
		var reqSchemaRef *openapi3.SchemaRef
		var resSchemaRef *openapi3.SchemaRef

		reqSchemaRef, err = generator.GenerateSchemaRef(info.ReqType)
		if err != nil {
			return nil, fmt.Errorf("failed to generate schema for request type %s: %w", info.ReqType.Name(), err)
		}
		
		resSchemaRef, err = generator.GenerateSchemaRef(info.ResType)
		if err != nil {
			return nil, fmt.Errorf("failed to generate schema for response type %s: %w", info.ResType.Name(), err)
		}

		operation := &openapi3.Operation{
			Summary:     fmt.Sprintf("Handle %s", info.ReqName),
			Description: fmt.Sprintf("Handles the %s command", info.ReqName),
			OperationID: info.ReqName,
		}

		operation.RequestBody = &openapi3.RequestBodyRef{
			Value: &openapi3.RequestBody{
				Required: true,
				Content:  openapi3.NewContentWithJSONSchema(reqSchemaRef.Value),
			},
		}
		operation.AddResponse(200, openapi3.NewResponse().
			WithContent(openapi3.NewContentWithJSONSchema(resSchemaRef.Value)))

		pathItem := &openapi3.PathItem{
			Post: operation,
		}

		paths.Set(fmt.Sprintf("/%s", info.ReqName), pathItem)
	}
	return paths, nil
}
