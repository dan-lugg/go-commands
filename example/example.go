package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"go-commands-v2/commands"
	"go-commands-v2/openapi"
	"io"
	"log"
	"net/http"
	"strings"
)

type AddCommandRes struct {
	Result int `json:"result"`
}

type AddCommandReq struct {
	ArgX int `json:"argX"`
	ArgY int `json:"argY"`
}

type AddHandler struct {
	commands.Handler[AddCommandReq, AddCommandRes]
}

func (h *AddHandler) Handle(req AddCommandReq, ctx context.Context) (res AddCommandRes, err error) {
	result := req.ArgX + req.ArgY
	return AddCommandRes{Result: result}, nil
}

type SubCommandRes struct {
	Result int `json:"result"`
}

type SubCommandReq struct {
	ArgX int `json:"argX"`
	ArgY int `json:"argY"`
}

type SubHandler struct {
	commands.Handler[SubCommandReq, SubCommandRes]
}

func (h *SubHandler) Handle(req SubCommandReq, ctx context.Context) (res SubCommandRes, err error) {
	result := req.ArgX - req.ArgY
	return SubCommandRes{Result: result}, nil
}

func main() {
	decoderRegistry := commands.NewDecoderRegistry()
	commands.RegisterDecoder[AddCommandReq](decoderRegistry, "add", commands.DefaultCommandReqDecoder[AddCommandReq]())
	commands.RegisterDecoder[SubCommandReq](decoderRegistry, "sub", commands.DefaultCommandReqDecoder[SubCommandReq]())

	handlerRegistry := commands.NewHandlerRegistry()
	commands.RegisterHandler[AddCommandReq, AddCommandRes](handlerRegistry, func() commands.Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})
	commands.RegisterHandler[SubCommandReq, SubCommandRes](handlerRegistry, func() commands.Handler[SubCommandReq, SubCommandRes] {
		return &SubHandler{}
	})

	http.HandleFunc("/commands/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/commands/")
		reqData, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("error reading request body: %v", err), http.StatusBadRequest)
			return
		}

		req, err := decoderRegistry.Decode(name, reqData)
		if err != nil {
			http.Error(w, fmt.Sprintf("error decoding request: %v", err), http.StatusBadRequest)
			return
		}

		res, err := handlerRegistry.Handle(req, context.Background())
		if err != nil {
			http.Error(w, fmt.Sprintf("error handling request: %v", err), http.StatusInternalServerError)
			return
		}

		resData, err := json.Marshal(res)
		if err != nil {
			http.Error(w, fmt.Sprintf("error encoding response: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(resData)
	})

	http.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		infos := handlerRegistry.Infos()
		paths, err := openapi.CreateOpenAPI3Paths(infos)
		if err != nil {
			http.Error(w, fmt.Sprintf("error creating OpenAPI paths: %v", err), http.StatusInternalServerError)
			return
		}

		swagger := openapi3.T{
			OpenAPI: "3.0.0",
			Info: &openapi3.Info{
				Title:       "Command API",
				Description: "API for handling commands",
				Version:     "1.0.0",
			},
			Paths: paths,
		}
		data, err := swagger.MarshalJSON()
		if err != nil {
			http.Error(w, fmt.Sprintf("error encoding response: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(data)
	})

	port := ":8080"
	println(fmt.Sprintf("Server listening on port %s...", port))
	log.Fatal(http.ListenAndServe(port, nil))
}
