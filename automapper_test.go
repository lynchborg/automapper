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
	c := New[from, to]().
		ForField("Struct", MapField(func(src from) (any, error) {
			return New[sub1, sub2]().Map(src.Struct)
		})).
		ForField("StructSlice", MapField(func(src from) (any, error) {
			return New[sub1, sub2]().MapSlice(src.StructSlice)
		})).
		ForField("Missing", IgnoreField()).
		ForField("WrongTypeSlice", MapField(func(src from) (any, error) {
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
	dest, err := c.Map(src)
	require.NoError(t, err)
	t.Log(dest)

}

func TestMissing(t *testing.T) {

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

func TestDifferentStructPointer(t *testing.T) {

	type sub1 struct{ Bar string }
	type sub2 struct{ Bar string }

	type from struct {
		Foo *sub1
	}

	type to struct {
		Foo *sub2
	}

	c := New[from, to]().
		ForField("Foo", MapField(func(s from) (any, error) {
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

func TestSameStructPointer(t *testing.T) {

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

func ref[T any](v T) *T {
	return &v
}
