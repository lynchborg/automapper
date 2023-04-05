# Automapper

Easily map from a source to a destination struct.

Can automatically map fields that are the same name and type.

Allows for custom override for fields that have differing types or that don't exist on the source struct.

```golang
package main

import (
	"fmt"
	"strconv"

	"github.com/lynchborg/automapper"
)

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

var c automapper.Config[from, to]

func init() {
	subMapper := automapper.New[sub1, sub2]()
	// first set up the mapper, advised to do this at startup once since it's costly with reflection
	// via string names
	c = automapper.New[from, to]().
		ForFieldName("Struct", automapper.MapField(func(src from) (any, error) {
			return subMapper.Map(src.Struct)
		})).
		ForFieldName("StructSlice", automapper.MapField(func(src from) (any, error) {
			return subMapper.MapSlice(src.StructSlice)
		})).
		ForFieldName("Missing", automapper.IgnoreField()).
		ForFieldName("WrongTypeSlice", automapper.MapField(func(src from) (any, error) {
			return automapper.MapSlice(src.WrongTypeSlice, func(input int) string {
				return strconv.Itoa(input)
			}), nil
		}))

	// or by func
	c = automapper.New[from, to]().
		ForField(func(dest *to) any {
			return &dest.Struct
		}, automapper.MapField(func(src from) (any, error) {
			return subMapper.Map(src.Struct)
		})).
		ForField(func(dest *to) any {
			return &dest.StructSlice
		}, automapper.MapField(func(src from) (any, error) {
			return subMapper.MapSlice(src.StructSlice)
		})).
		ForField(func(dest *to) any {
			return &dest.Missing
		}, automapper.IgnoreField()).
		ForField(func(dest *to) any {
			return &dest.WrongTypeSlice
		}, automapper.MapField(func(src from) (any, error) {
			return automapper.MapSlice(src.WrongTypeSlice, func(input int) string {
				return strconv.Itoa(input)
			}), nil
		}))

}

func main() {

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
	// do your mapping
	dest, err := c.Map(src)
	if err != nil {
		panic(err)
	}

	fmt.Println(dest)
}

```
