package commands

import "testing"

func Test_Full(t *testing.T) {
	decoderCatalog := NewDefaultDecoderCatalog()
	InsertDecoder[AddCommandReq](decoderCatalog, NewDefaultDecoder[AddCommandReq]())
	InsertDecoder[SubCommandReq](decoderCatalog, NewDefaultDecoder[SubCommandReq]())

	mappingCatalog := NewDefaultMappingCatalog()
	InsertMapping[AddCommandReq](mappingCatalog, AddReqName)
	InsertMapping[SubCommandReq](mappingCatalog, SubReqName)

	handlerCatalog := NewDefaultHandlerCatalog()
	InsertHandler[AddCommandReq, AddCommandRes](handlerCatalog, func() Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})
	InsertHandler[SubCommandReq, SubCommandRes](handlerCatalog, func() Handler[SubCommandReq, SubCommandRes] {
		return &SubHandler{}
	})

	reqName := AddReqName
	reqType, err := mappingCatalog.ByName(reqName)
	if err != nil {
		t.Fatalf("no reqType found for reqName %s: %v", reqName, err)
	}

	req, err := decoderCatalog.Decode(reqType, []byte(`{"argX":3,"argY":4}`))
	if err != nil {
		t.Fatalf("no req found for req name %s: %v", reqType, err)
	}

	res, err := handlerCatalog.Handle(nil, req)
	if err != nil {
		t.Fatalf("handler error for req name %s: %v", reqType, err)
	}

	addRes, ok := res.(AddCommandRes)
	if !ok {
		t.Fatalf("unexpected response type: %T", res)
	}
	expected := AddCommandRes{Result: 7}
	if addRes != expected {
		t.Fatalf("unexpected response value: got %v, want %v", addRes, expected)
	}
}
