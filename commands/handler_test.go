package commands

import (
	"context"
	"fmt"
	"github.com/dan-lugg/go-commands/futures"
	"github.com/dan-lugg/go-commands/util"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func Test_NewDefaultHandlerAdapter(t *testing.T) {
	adapter := NewDefaultHandlerAdapter(func() Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})
	assert.NotNil(t, adapter)
	assert.IsType(t, &DefaultHandlerAdapter[AddCommandReq, AddCommandRes]{}, adapter)
}

func Test_DefaultHandlerAdapter_Handle(t *testing.T) {
	adapter := NewDefaultHandlerAdapter(func() Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})

	t.Run("valid req", func(t *testing.T) {
		res, err := adapter.Handle(nil, AddCommandReq{ArgX: 3, ArgY: 4})
		assert.NoError(t, err)
		assert.Equal(t, AddCommandRes{Result: 7}, res)
	})

	t.Run("invalid req", func(t *testing.T) {
		res, err := adapter.Handle(nil, SubCommandReq{})
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func Test_DefaultHandlerAdapter_ReqType(t *testing.T) {
	adapter := NewDefaultHandlerAdapter(func() Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})
	assert.Equal(t, reflect.TypeFor[AddCommandReq](), adapter.ReqType())
}

func Test_DefaultHandlerAdapter_ResType(t *testing.T) {
	adapter := NewDefaultHandlerAdapter(func() Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})
	assert.Equal(t, reflect.TypeFor[AddCommandRes](), adapter.ResType())
}

func Test_NewHandlerCatalog(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		catalog := NewHandlerCatalog()
		assert.NotNil(t, catalog)
		assert.Empty(t, catalog.adapters)
		assert.IsType(t, &HandlerCatalog{}, catalog)
	})

	t.Run("with options", func(t *testing.T) {
		catalog := NewHandlerCatalog(func(*HandlerCatalog) {})
		assert.NotNil(t, catalog)
		assert.Empty(t, catalog.adapters)
		assert.IsType(t, &HandlerCatalog{}, catalog)
	})
}

func Test_HandlerCatalog_Insert(t *testing.T) {
	t.Run("empty catalog", func(t *testing.T) {
		catalog := HandlerCatalog{}
		assert.Nil(t, catalog.adapters)
		adapter := NewDefaultHandlerAdapter(func() Handler[AddCommandReq, AddCommandRes] {
			return &AddHandler{}
		})
		catalog.Insert(adapter)
		assert.NotEmpty(t, catalog.adapters)
		assert.Contains(t, catalog.adapters, adapter.ReqType())
	})

	t.Run("constructed catalog", func(t *testing.T) {
		catalog := NewHandlerCatalog()
		assert.NotNil(t, catalog)
		adapter := NewDefaultHandlerAdapter(func() Handler[AddCommandReq, AddCommandRes] {
			return &AddHandler{}
		})
		catalog.Insert(adapter)
		assert.NotEmpty(t, catalog.adapters)
		assert.Contains(t, catalog.adapters, adapter.ReqType())
	})
}

func Test_InsertHandler(t *testing.T) {
	catalog := NewHandlerCatalog()
	InsertHandler[AddCommandReq, AddCommandRes](catalog, func() Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})
	assert.NotEmpty(t, catalog.adapters)
	assert.Contains(t, catalog.adapters, reflect.TypeFor[AddCommandReq]())
}

func Test_HandlerCatalog_Handle(t *testing.T) {
	catalog := NewHandlerCatalog()
	InsertHandler[AddCommandReq, AddCommandRes](catalog, func() Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})

	t.Run("default", func(t *testing.T) {
		res, err := Handle[AddCommandReq, AddCommandRes](nil, catalog, AddCommandReq{ArgX: 3, ArgY: 4})
		assert.NoError(t, err)
		assert.Equal(t, AddCommandRes{Result: 7}, res)
	})

	t.Run("handler missing", func(t *testing.T) {
		res, err := Handle[SubCommandReq, SubCommandRes](nil, catalog, SubCommandReq{ArgX: 3, ArgY: 4})
		assert.Zero(t, res)
		assert.ErrorIs(t, err, ErrHandlerMissing)
	})
}

type SlowCommandRes struct {
	CommandRes
	Name string
}
type SlowCommandReq struct {
	CommandReq[SlowCommandRes]
	Name string
	Iter int
}

type SlowHandler struct {
	Handler[SlowCommandReq, SlowCommandRes]
}

func (h *SlowHandler) Handle(ctx context.Context, req SlowCommandReq) (res SlowCommandRes, err error) {
	for i := 1; i <= req.Iter; i++ {
		println("Processing in SlowHandler:", i, "for command:", req.Name)
		time.Sleep(1 * time.Second)
	}
	return SlowCommandRes{
		Name: req.Name,
	}, nil
}

func Test_HandlerCatalog_Future(t *testing.T) {
	catalog := NewHandlerCatalog()
	InsertHandler[AddCommandReq, AddCommandRes](catalog, func() Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})
	InsertHandler[SlowCommandReq, SlowCommandRes](catalog, func() Handler[SlowCommandReq, SlowCommandRes] {
		return &SlowHandler{}
	})

	t.Run("default", func(t *testing.T) {
		fut := Future[AddCommandReq, AddCommandRes](nil, catalog, AddCommandReq{ArgX: 3, ArgY: 4})
		tup := fut.Wait()
		res, err := tup.Val1, tup.Val2

		assert.Equal(t, AddCommandRes{Result: 7}, res)
		assert.NoError(t, err)
	})

	t.Run("handler missing", func(t *testing.T) {
		fut := Future[SubCommandReq, SubCommandRes](nil, catalog, SubCommandReq{ArgX: 3, ArgY: 4})
		tup := fut.Wait()
		res, err := tup.Val1, tup.Val2
		assert.Zero(t, res)
		assert.ErrorIs(t, err, ErrHandlerMissing)
	})

	t.Run("multiple", func(t *testing.T) {
		start := time.Now()

		fut1 := Future[SlowCommandReq, SlowCommandRes](nil, catalog, SlowCommandReq{
			Name: "A",
			Iter: 1,
		})
		fut2 := Future[SlowCommandReq, SlowCommandRes](nil, catalog, SlowCommandReq{
			Name: "B",
			Iter: 3,
		})
		fut3 := Future[SlowCommandReq, SlowCommandRes](nil, catalog, SlowCommandReq{
			Name: "C",
			Iter: 4,
		})
		fut4 := Future[SlowCommandReq, SlowCommandRes](nil, catalog, SlowCommandReq{
			Name: "D",
			Iter: 2,
		})

		tups := futures.WaitAll[util.Tuple2[SlowCommandRes, error]](fut1, fut2, fut3, fut4).Wait()

		duration := time.Since(start)

		assert.Less(t, duration, 5*time.Second)
		assert.Greater(t, duration, 3*time.Second)

		for _, tup := range tups {
			res, err := tup.Val1, tup.Val2
			assert.NoError(t, err)
			assert.IsType(t, SlowCommandRes{}, res)
			println(fmt.Sprintf("Processed command: %s", res.Name))
		}
	})
}

func Test_HandlerCatalog_TypeMap(t *testing.T) {
	catalog := NewHandlerCatalog()
	InsertHandler[AddCommandReq, AddCommandRes](catalog, func() Handler[AddCommandReq, AddCommandRes] {
		return &AddHandler{}
	})
	InsertHandler[SubCommandReq, SubCommandRes](catalog, func() Handler[SubCommandReq, SubCommandRes] {
		return &SubHandler{}
	})
	typeMap := catalog.TypeMap()
	assert.NotEmpty(t, typeMap)
	assert.Contains(t, typeMap, reflect.TypeFor[AddCommandReq]())
	assert.Contains(t, typeMap, reflect.TypeFor[SubCommandReq]())
	assert.Equal(t, reflect.TypeFor[AddCommandRes](), typeMap[reflect.TypeFor[AddCommandReq]()])
	assert.Equal(t, reflect.TypeFor[SubCommandRes](), typeMap[reflect.TypeFor[SubCommandReq]()])
}
