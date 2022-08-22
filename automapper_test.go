package automapper

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestM(t *testing.T) {
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
			var ret []string
			for _, item := range src.WrongTypeSlice {
				ret = append(ret, strconv.Itoa(item))
			}
			return ret, nil
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
