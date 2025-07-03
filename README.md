# Go Commands

![CI Build](https://github.com/dan-lugg/go-commands/actions/workflows/ci.yml/badge.svg)

This repository provides a framework for handling commands in Go, including request decoding, handler registration, and
execution. It simplifies the process of managing command requests and responses in a structured and extensible way.

## Features

- **Command Request and Response Interfaces**: Define generic interfaces for command requests and responses.
- **Decoder Registry**: Manage mappings between request names, types, and decoders for serialized data.
- **Handler Registry**: Register and manage handlers for processing command requests.
- **Default Handler Adapter**: Adapt handlers to a common structure for consistent processing.
- **JSON Resolver**: Resolve JSON input into request names and data.

## Installation

Ensure you have Go installed. Clone the repository and run:

```bash
go mod tidy
```

This will install all required dependencies.

## Usage

### Command Request and Response

Define your command request and response types by implementing the `CommandReq` and `CommandRes` interfaces.

```go
type AddCommandReq struct {
	ArgX int `json:"argX"`
	ArgY int `json:"argY"`
}

type AddCommandRes struct {
	Result int `json:"result"`
}

type SubCommandReq struct {
	ArgX int `json:"argX"`
	ArgY int `json:"argY"`
}

type SubCommandRes struct {
	Result int `json:"result"`
}
```

### Registering Decoders

Use the `DecoderRegistry` to register decoders for your command request types.

```go
decoderRegistry := NewDecoderRegistry()
RegisterDecoder[AddCommandReq](decoderRegistry, "add", DefaultCommandReqDecoder[AddCommandReq]())
RegisterDecoder[SubCommandReq](decoderRegistry, "sub", DefaultCommandReqDecoder[SubCommandReq]())
```

### Registering Handlers

Define handlers for your command requests by implementing the `Handler` interface.

```go
type AddHandler struct {
    commands.Handler[AddCommandReq, AddCommandRes]
}

func (h *AddHandler) Handle(req AddCommandReq, ctx context.Context) (res AddCommandRes, err error) {
    result := req.ArgX + req.ArgY
    return AddCommandRes{Result: result}, nil
}

type SubHandler struct {
    commands.Handler[SubCommandReq, SubCommandRes]
}

func (h *SubHandler) Handle(req SubCommandReq, ctx context.Context) (res SubCommandRes, err error) {
    result := req.ArgX - req.ArgY
    return SubCommandRes{Result: result}, nil
}
```

Use the `HandlerRegistry` to register handlers for your command request types.

```go
handlerRegistry := NewHandlerRegistry()
RegisterHandler[AddCommandReq, AddCommandRes](handlerRegistry, func () Handler[AddCommandReq, AddCommandRes] {
    return &AddHandler{}
})
RegisterHandler[SubCommandReq, SubCommandRes](handlerRegistry, func () Handler[SubCommandReq, SubCommandRes] {
    return &SubHandler{}
})
```

### Handling Requests

Use the `HandlerRegistry` to process command requests.

```go
res, err := handlerRegistry.Handle(AddCommandReq{ArgX: 5, ArgY: 3}, context.Background())
if err != nil {
    log.Fatalf("Error handling request: %v", err)
}
fmt.Printf("Response: %+v\n", res)
```

## Testing

Run unit tests using:

```bash
go test ./...
```

## Project Structure

- `commands/commands.go`: Core framework implementation.
- `commands/commands_test.go`: Unit tests for the framework.
- `example/example.go`: Example usage of the framework.

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