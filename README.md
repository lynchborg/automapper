# Automapper

Easily map from a source to a destination struct.
Can automatically map fields that are the same name and type.
Allows for custom override for fields that have differing types or that don't exist on the source struct.

```golang

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
type sub1 struct{ Bar string }

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
type sub2 struct{ Bar string }

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
//
// outputs:
// to{
//     String:      "String",
//     Int:         999,
//     Struct:      sub1{Bar: "hey"},
//     StringSlice: []string{"1", "2"},
//     StructSlice: []sub2{
//     {Bar: "1"},
//     {Bar: "2"},
//     },
//     WrongTypeSlice: []string{"999","998"},
// }
```
