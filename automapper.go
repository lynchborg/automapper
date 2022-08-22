package automapper

import (
	"fmt"
	"reflect"
)

type Config[S any, D any] struct {
	fieldMappings map[string]*Opts
}

func New[S any, D any]() Config[S, D] {
	m := Config[S, D]{
		fieldMappings: map[string]*Opts{},
	}
	return m
}

type IncompatibleTypesErr struct {
	src  reflect.Type
	dest reflect.Type
}

func (i IncompatibleTypesErr) Error() string {
	return fmt.Sprintf("destination type is %s, source is %s", i.dest, i.src)
}

func (m Config[S, D]) mapAny(srcType reflect.Type, srcValue reflect.Value, destType reflect.Type, destValue reflect.Value) error {

	switch destType.Kind() {
	case reflect.Struct:
		if srcType.Kind() != reflect.Struct {
			return IncompatibleTypesErr{src: srcType, dest: destType}
		}
		if srcType == destType {
			destValue.Set(srcValue)
			return nil
		}

	case reflect.Slice:
		if srcType.Elem() != destType.Elem() {
			return IncompatibleTypesErr{src: srcType, dest: destType}
		}
		destValue.Set(srcValue)
	case reflect.Pointer:
	default:
		if srcType != destType {
			return IncompatibleTypesErr{src: srcType, dest: destType}
		}
		destValue.Set(srcValue)
	}
	return nil
}

func (m Config[S, D]) mapField(src S, srcField reflect.StructField, srcFieldValue reflect.Value, destField reflect.StructField, destFieldValue reflect.Value) error {

	return m.mapAny(srcField.Type, srcFieldValue, destField.Type, destFieldValue)

}

func (m Config[S, D]) MapSlice(src []S) ([]D, error) {

	var ret []D
	for _, item := range src {
		mappedItem, err := m.Map(item)
		if err != nil {
			return nil, err
		}
		ret = append(ret, mappedItem)
	}
	return ret, nil
}

func (m Config[S, D]) Map(src S) (D, error) {
	// loop through fields with reflection
	// check if field in our map
	dest := new(D)
	srcType := reflect.TypeOf(src)
	srcNumField := srcType.NumField()
	srcValue := reflect.ValueOf(&src)

	destType := reflect.TypeOf(*dest)
	destNumField := destType.NumField()
	destValue := reflect.ValueOf(dest)

	for j := 0; j < destNumField; j++ {
		destFieldValue := destValue.Elem().Field(j)
		destFieldType := destType.Field(j)
		found := false
		fieldMapping, ok := m.fieldMappings[destFieldType.Name]
		if ok {
			err := fieldMapping.apply(src, destFieldValue)
			if err != nil {
				return *dest, err
			}
			continue
		}
		for i := 0; i < srcNumField; i++ {
			srcFieldValue := srcValue.Elem().Field(i)
			srcFieldType := srcType.Field(i)
			if destFieldType.Name == srcFieldType.Name {
				found = true
				if err := m.mapField(src, srcFieldType, srcFieldValue, destFieldType, destFieldValue); err != nil {
					return *dest, err
				}
				break
			}
		}
		if !found {
			return *dest, fmt.Errorf("field '%s' not found in source type '%s'", destFieldType.Name, reflect.TypeOf(src))

		}
	}

	return *dest, nil
}

type Opts struct {
	mapFunc func(any) (any, error)
	ignore  bool
}

func (o Opts) apply(src any, destValue reflect.Value) error {
	if o.ignore {
		return nil
	}
	if o.mapFunc != nil {
		v, err := o.mapFunc(src)
		if err != nil {
			return err
		}
		destValue.Set(reflect.ValueOf(v))
	}
	return nil
}
func IgnoreField() func(o *Opts) {
	return func(o *Opts) {
		o.ignore = true
	}
}

func MapField[S any](mapFunc func(S) (any, error)) func(o *Opts) {
	return func(o *Opts) {
		o.mapFunc = func(s any) (any, error) {
			return mapFunc(s.(S))
		}
	}
}
func (m Config[S, D]) ForField(name string, option func(o *Opts)) Config[S, D] {
	dest := new(D)
	_, found := reflect.TypeOf(*dest).FieldByName(name)
	if !found {
		panic(fmt.Errorf("destination has no field named %s", name))
	}
	opts, found := m.fieldMappings[name]
	if !found {
		opts = &Opts{}
	}
	// run now early
	option(opts)
	m.fieldMappings[name] = opts
	return m
}
