package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrDecoderMissing = errors.New("decoder missing")
	ErrDecoderFailure = errors.New("decoder failure")
)

// handler
// decoder

type CommandRes any

type CommandReq[TRes CommandRes] any

type Handler[TReq CommandReq[TRes], TRes CommandRes] interface {
	Handle(ctx context.Context, req TReq) (res TRes, err error)
}

type Decoder interface {
	Decode(reqData []byte) (req CommandReq[CommandRes], err error)
}

type CompositeDecoder struct {
	decoders []Decoder
}

func NewCompositeDecoder() *CompositeDecoder {
	return &CompositeDecoder{
		decoders: make([]Decoder, 0),
	}
}

func (c *CompositeDecoder) AddDecoder(decoder Decoder) {
	c.decoders = append(c.decoders, decoder)
}

func (c *CompositeDecoder) Decode(reqData []byte) (req CommandReq[CommandRes], err error) {
	if len(c.decoders) == 0 {
		return nil, fmt.Errorf("%w: no decoders registered", ErrDecoderMissing)
	}
	for _, decoder := range c.decoders {
		req, err = decoder.Decode(reqData)
		if err != nil {
			continue
		}
		return req, nil
	}
	return nil, fmt.Errorf("%w: all decoders failed", ErrDecoderFailure)
}

type JSONDecoder struct{}

func (d *JSONDecoder) Decode(reqData []byte) (req CommandReq[CommandRes], err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	err = json.Unmarshal(reqData, &req)
	if err != nil {
		var syntaxErr *json.SyntaxError
		if errors.As(err, &syntaxErr) {
			return nil, fmt.Errorf("%w: at byte offset %d", err, syntaxErr.Offset)
		}
		var unmarshalTypeErr *json.UnmarshalTypeError
		if errors.As(err, &unmarshalTypeErr) {
			return nil, fmt.Errorf("%w: at byte offset %d", err, unmarshalTypeErr.Offset)
		}
	}
	return req, nil
}

type YAMLDecoder struct{}

func (d *YAMLDecoder) Decode(reqData []byte) (req CommandReq[CommandRes], err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	return nil, fmt.Errorf("YAMLDecoder not implemented")
}
