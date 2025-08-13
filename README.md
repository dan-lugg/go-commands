# Go Commands

![CI Build](https://github.com/dan-lugg/go-commands/actions/workflows/ci.yml/badge.svg)
![GitHub Tag](https://img.shields.io/github/v/tag/dan-lugg/go-commands?style=flat)

This repository provides a framework for handling commands in Go, including request decoding, handler registration, and
execution. It simplifies the process of managing command requests and results in a structured and extensible way.

## Features

- **Command Request and Result Interfaces**:
    - Define generic interfaces for command requests and results, enabling type-safe and reusable command structures.
- **Handler Catalog**:
    - Register and manage handlers for processing command requests, ensuring modular and extensible command execution.
- **Mapping Catalog**:
    - Map request names to types for easier integration with external systems.
- **Decoder Catalog**:
    - Manage mappings between request types and decoders for serialized data, allowing flexible deserialization of
      incoming requests.
- **Future Support**: Asynchronous processing of commands using `futures.Future`.

## Installation

Ensure you have Go installed. Clone the repository and run:

```bash
go get github.com/dan-lugg/go-commands
```

This will install all required dependencies and prepare the project for development.

## Usage

### Defining Commands

Define your command request and result types by implementing the `CommandReq` and `CommandRes` interfaces. These
interfaces ensure that your commands are structured and type-safe.

```go
package example

import "github.com/dan-lugg/go-commands/commands"

// AddCommandRes represents the result for an Add command.
type AddCommandRes struct {
	// Embed the CommandRes interface to ensure compatibility with the framework.
	commands.CommandRes

	// Result of the Add command.
	Result int
}

// AddCommandReq represents the request for an Add command.
type AddCommandReq struct {
	// Embed the CommandReq interface to ensure compatibility with the framework.
	commands.CommandReq[AddCommandReq]

	// Arguments for the Add command.
	ArgX int
	ArgY int
}

```

### Registering Handlers

Define handlers for your command requests by implementing the `Handler` interface. Handlers process the command requests
and return the corresponding results.

```go
package example

import "context"

// AddHandler processes AddCommandReq and returns AddCommandRes.
type AddHandler struct {
	// Embed the Handler interface to ensure compatibility with the framework.
	Handler[AddCommandReq, AddCommandRes]
}

// Handle processes the AddCommandReq and returns an AddCommandRes.
func (h *AddHandler) Handle(ctx context.Context, req AddCommandReq) (AddCommandRes, error) {
	return AddCommandRes{Result: req.ArgX + req.ArgY}, nil
}

```

Register these handlers using the `HandlerCatalog`.

```go
package example

import "github.com/dan-lugg/go-commands/commands"

func exampleHandlerRegistration() {
	// Create a new HandlerCatalog
	handlerCatalog := commands.NewHandlerCatalog()

	// Register the AddHandler
	commands.InsertHandler[AddCommandReq, AddCommandRes](
		handlerCatalog,
		func() Handler[AddCommandReq, AddCommandRes] {
			return &AddHandler{}
		},
	)

	// Register the SubHandler
	commands.InsertHandler(
		handlerCatalog,
		func() commands.Handler[SubCommandReq, SubCommandRes] {
			return &SubHandler{}
		},
	)
}

```

### Handling Requests

Use the `HandlerCatalog` to process command requests. The catalog will route the request to the appropriate handler
based on its type.

```go
package example

import (
	"fmt"
	"github.com/dan-lugg/go-commands/commands"
	"log"
)

func exampleRequestHandling() {
	// Create a new HandlerCatalog
	handlerCatalog := commands.NewHandlerCatalog()

	// Register handlers
	exampleHandlerRegistration()

	// Create a command request
	req := AddCommandReq{ArgX: 5, ArgY: 3}

	// Handle the request
	res, err := commands.Handle[AddCommandReq, AddCommandRes](context.Background(), handlerCatalog, req)
	if err != nil {
		log.Fatalf("error handling request: %v", err)
	}

	fmt.Printf("result: %+v\n", res)
}

```

### Asynchronous Processing

The `Future` function allows you to process commands asynchronously.

```go
package example

func exampleAsyncProcessing() {
	// Create a new HandlerCatalog
	handlerCatalog := commands.NewHandlerCatalog()

	// Register handlers
	exampleHandlerRegistration()

	// Use Future to handle the command asynchronously
	future := commands.Future[AddCommandReq, AddCommandRes](context.Background(), handlerCatalog, AddCommandReq{ArgX: 5, ArgY: 3})

	// Wait for the result
	result := future.Wait()
	fmt.Printf("async result: %+v\n", result.Val1)
}
```

### Registering Mappers

Use the `MappingCatalog` to map request names to their corresponding types.

```go
package example

import (
	"fmt"
	"github.com/dan-lugg/go-commands/commands"
	"log"
)

func exampleMappingCatalog() {
	// Create a new MappingCatalog
	mappingCatalog := commands.NewMappingCatalog()

	// Insert mappings for command request types
	commands.InsertMapping[AddCommandReq](mappingCatalog, "add")
	commands.InsertMapping[SubCommandReq](mappingCatalog, "sub")

	// Retrieve a type by its name
	reqType, err := mappingCatalog.ByName("add")
	if err != nil {
		log.Fatalf("error retrieving type: %v", err)
	}
	fmt.Printf("request type: %v\n", reqType)
}

```

### Registering Decoders

Use the `DecoderCatalog` to register decoders for your command request types. Decoders are responsible for deserializing
incoming data into specific command request types.

```go
package example

import "github.com/dan-lugg/go-commands/commands"

func exampleDecoderRegistration() {
	// Create a new DecoderCatalog
	decoderCatalog := commands.NewDecoderCatalog()

	// Register decoders for command requests
	commands.InsertDecoder[AddCommandReq](decoderCatalog, DefaultDecoder[AddCommandReq]())
	commands.InsertDecoder[SubCommandReq](decoderCatalog, DefaultDecoder[SubCommandReq]())
}

```

## Testing

Unit tests are provided to ensure the reliability of the framework. Run the tests using:

```bash
go test ./...
```

## Project Structure

- `commands/`:
    - Core framework implementation.
- `futures/`:
    - Asynchronous processing utilities.
- `util/`:
    - Utility types and functions.

## Dependencies

- [Testify](https://github.com/stretchr/testify): For assertions in unit tests.

## Contributing

Contributions are welcome! Follow these steps:

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Submit a pull request with a clear description of your changes.

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.

## Contact

For questions or feedback, open an issue in the repository.