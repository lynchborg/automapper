package automapper

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMainFeatures(t *testing.T) {
	type sub1 struct{ Bar string }
	type sub2 struct{ Bar string }

	type from struct {
		String string
		Int    int
		Inline struct {
			Foo string
		}
		Struct         sub1
		StringSlice    []string
		StructSlice    []sub1
		WrongTypeSlice []int
	}

	type to struct {
		Int    int
		String string
		Inline struct {
			Foo string
		}
		Struct         sub2
		StringSlice    []string
		StructSlice    []sub2
		WrongTypeSlice []string
		Missing        bool
	}

	dest := to{}

	src := from{
		String:      "String",
		Int:         999,
		Struct:      sub1{Bar: "hey"},
		StringSlice: []string{"1", "2"},
		StructSlice: []sub1{
			{Bar: "1"},
			{Bar: "2"},
		},
		WrongTypeSlice: []int{999, 998},
	}

	c := New[from, to]().
		ForFieldName("Struct", MapField(func(src from) (any, error) {
			return New[sub1, sub2]().Map(src.Struct)
		})).
		ForFieldName("StructSlice", MapField(func(src from) (any, error) {
			return New[sub1, sub2]().MapSlice(src.StructSlice)
		})).
		ForFieldName("Missing", IgnoreField()).
		ForFieldName("WrongTypeSlice", MapField(func(src from) (any, error) {
			return MapSlice(src.WrongTypeSlice, func(input int) string {
				return strconv.Itoa(input)
			}), nil
		}))
	dest, err := c.Map(src)
	require.NoError(t, err)

	c = New[from, to]().
		ForField(func(dest *to) any {
			return &dest.Struct
		}, MapField(func(src from) (any, error) {
			return New[sub1, sub2]().Map(src.Struct)
		})).
		ForField(func(dest *to) any {
			return &dest.StructSlice
		}, MapField(func(src from) (any, error) {
			return New[sub1, sub2]().MapSlice(src.StructSlice)
		})).
		ForField(func(dest *to) any {
			return &dest.Missing
		}, IgnoreField()).
		ForField(func(dest *to) any {
			return &dest.WrongTypeSlice
		}, MapField(func(src from) (any, error) {
			return MapSlice(src.WrongTypeSlice, func(input int) string {
				return strconv.Itoa(input)
			}), nil
		}))

	dest, err = c.Map(src)
	require.NoError(t, err)
	t.Log(dest)

}

func BenchmarkNormalStruct(b *testing.B) {

	type sub1 struct{ Bar string }
	type sub2 struct{ Bar string }

	type from struct {
		String string
		Int    int
		Inline struct {
			Foo string
		}
		Struct         sub1
		StringSlice    []string
		StructSlice    []sub1
		WrongTypeSlice []int
	}

	type to struct {
		Int    int
		String string
		Inline struct {
			Foo string
		}
		Struct         sub2
		StringSlice    []string
		StructSlice    []sub2
		WrongTypeSlice []string
		Missing        bool
	}

	subConfig := New[sub1, sub2]()

	c := New[from, to]().
		ForField(func(dest *to) any {
			return &dest.Struct
		}, MapField(func(src from) (any, error) {
			return subConfig.Map(src.Struct)
		})).
		ForField(func(dest *to) any {
			return &dest.StructSlice
		}, MapField(func(src from) (any, error) {
			return subConfig.MapSlice(src.StructSlice)
		})).
		ForField(func(dest *to) any {
			return &dest.Missing
		}, IgnoreField()).
		ForField(func(dest *to) any {
			return &dest.WrongTypeSlice
		}, MapField(func(src from) (any, error) {
			return MapSlice(src.WrongTypeSlice, func(input int) string {
				return strconv.Itoa(input)
			}), nil
		}))

	src := from{
		String:      "String",
		Int:         999,
		Struct:      sub1{Bar: "hey"},
		StringSlice: []string{"1", "2"},
		StructSlice: []sub1{
			{Bar: "1"},
			{Bar: "2"},
		},
		WrongTypeSlice: []int{999, 998},
	}

	_, err := c.Map(src)
	require.NoError(b, err)

	b.Run("using automapper", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = c.Map(src)
		}
	})

	b.Run("manually copied", func(b *testing.B) {
		_ = &to{
			Int:    src.Int,
			String: src.String,
			Inline: struct {
				Foo string
			}(src.Inline),
			Struct:      sub2{Bar: src.Struct.Bar},
			StringSlice: src.StringSlice,
			StructSlice: MapSlice(src.StructSlice, func(input sub1) sub2 {
				return sub2{Bar: input.Bar}
			}),
			WrongTypeSlice: MapSlice(src.WrongTypeSlice, func(input int) string {
				return strconv.Itoa(input)
			}),
			Missing: false,
		}

	})

}

func TestShouldRespectMissing(t *testing.T) {

	type from struct {
		Foo string
	}

	type to struct {
		Missing bool
	}

	c := New[from, to]()
	_, err := c.Map(from{Foo: "foo"})
	require.Error(t, err)
	require.Equal(t, "field 'Missing' not found in source type 'automapper.from'", err.Error())
}

func TestPrimitivePointer(t *testing.T) {

	type from struct {
		Foo *int
	}

	type to struct {
		Foo *int
	}

	c := New[from, to]()
	res, err := c.Map(from{Foo: ref(99)})
	require.NoError(t, err)
	require.NotNil(t, res.Foo)
	require.Equal(t, 99, *res.Foo)
}

func TestShouldMapDifferentStructPointer(t *testing.T) {

	type sub1 struct{ Bar string }
	type sub2 struct{ Bar string }

	type from struct {
		Foo *sub1
	}

	type to struct {
		Foo *sub2
	}

	c := New[from, to]().
		ForFieldName("Foo", MapField(func(s from) (any, error) {
			if s.Foo == nil {
				return nil, nil

			}
			return &sub2{Bar: s.Foo.Bar}, nil
		}))
	res, err := c.Map(from{Foo: &sub1{Bar: "bar"}})
	require.NoError(t, err)
	require.NotNil(t, res.Foo)
	require.Equal(t, sub2{Bar: "bar"}, *res.Foo)
}

func TestShouldMapSameStructPointer(t *testing.T) {

	type sub1 struct{ Bar string }

	type from struct {
		Foo *sub1
	}

	type to struct {
		Foo *sub1
	}

	c := New[from, to]()
	res, err := c.Map(from{Foo: &sub1{Bar: "bar"}})
	require.NoError(t, err)
	require.NotNil(t, res.Foo)
	require.Equal(t, sub1{Bar: "bar"}, *res.Foo)
}

func TestShouldMapCustomPrimitives(t *testing.T) {

	type custom string
	type from struct {
		Foo    string
		FooPtr *string
		Foos   []string
	}

	type to struct {
		Foo    custom
		FooPtr *custom
		Foos   []custom
	}

	c := New[from, to]()
	res, err := c.Map(from{Foo: "custom", FooPtr: ref("ptr"), Foos: []string{"one"}})
	require.NoError(t, err)
	require.Equal(t, custom("custom"), res.Foo)
	require.Equal(t, ref(custom("ptr")), res.FooPtr)
	require.Equal(t, []custom{"one"}, res.Foos)

}

func ref[T any](v T) *T {
	return &v
}

func TestShouldFailIfNotStruct(t *testing.T) {

	defer func() {
		if err := recover(); err != nil {
			t.Log(err)
			return
		}
		t.Fatal("should have panicked")
	}()
	_ = New[int, bool]()
}
