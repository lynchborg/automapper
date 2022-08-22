# Automapper

Easily map from a source to a destination struct.

Can automatically map fields that are the same name and type.

Allows for custom override for fields that have differing types or that don't exist on the source struct.

```golang

type from struct {
	...
}

type to struct {
	...
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
        return MapSlice(src.WrongTypeSlice, func(i int) string { return strconv.Itoa(i)})
    }))

src := from{
    ...
}
dest, err := c.Map(src)
```
