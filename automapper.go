package automapper

import (
	"fmt"
	"reflect"
)

type Config[S any, D any] struct {
	fieldMappings map[string]Opts
	destType      reflect.Type
	destFields    int
	srcType       reflect.Type
	srcFields     int
}

func New[S any, D any]() Config[S, D] {
	dest := new(D)
	src := new(S)
	destType := reflect.TypeOf(*dest)
	srcType := reflect.TypeOf(*src)
	m := Config[S, D]{
		fieldMappings: map[string]Opts{},
		destType:      destType,
		destFields:    destType.NumField(),
		srcType:       srcType,
		srcFields:     srcType.NumField(),
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
		if srcType != destType {
			return IncompatibleTypesErr{src: srcType, dest: destType}
		}
		destValue.Set(srcValue)
		return nil

	case reflect.Slice:
		if srcType.Kind() != reflect.Slice {
			return IncompatibleTypesErr{src: srcType, dest: destType}
		}

		if srcType.Elem().Kind() != destType.Elem().Kind() {
			return IncompatibleTypesErr{src: srcType, dest: destType}
		}
		if srcValue.IsNil() {
			return nil
		}
		if srcType.Elem() != destType.Elem() {
			// need to cast
			destValue.Set(reflect.MakeSlice(destType, srcValue.Len(), srcValue.Len()))
			for i := 0; i < srcValue.Len(); i++ {
				elemValue := srcValue.Index(i)
				destValue.Index(i).Set(elemValue.Convert(destType.Elem()))
			}
			return nil
		}
		destValue.Set(srcValue)
	case reflect.Pointer:
		referencedDestType := destType.Elem()
		referencedSourceType := srcType.Elem()
		if referencedSourceType.Kind() != referencedDestType.Kind() {
			return IncompatibleTypesErr{src: srcType, dest: destType}
		}

		if srcValue.IsNil() {
			return nil
		}

		if referencedSourceType == referencedDestType {
			destValue.Set(srcValue)
		}
		destValue.Set(srcValue.Convert(destValue.Type()))
	default:
		if srcType.Kind() != destType.Kind() {
			return IncompatibleTypesErr{src: srcType, dest: destType}
		}
		if destValue.Type() != srcValue.Type() {
			srcValue = srcValue.Convert(destValue.Type())
		}

		destValue.Set(srcValue)
	}
	return nil
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
	dest := new(D)

	srcValue := reflect.ValueOf(&src)
	srcType := m.srcType
	destValue := reflect.ValueOf(dest)
	destType := m.destType

	for j := 0; j < m.destFields; j++ {
		destFieldType := destType.Field(j)
		if !destFieldType.IsExported() {
			continue
		}

		destFieldValue := destValue.Elem().Field(j)
		found := false
		fieldMapping, ok := m.fieldMappings[destFieldType.Name]
		if ok {
			err := fieldMapping.apply(src, destFieldValue)
			if err != nil {
				return *dest, err
			}
			continue
		}
		for i := 0; i < m.srcFields; i++ {
			srcFieldValue := srcValue.Elem().Field(i)
			srcFieldType := srcType.Field(i)
			if destFieldType.Name == srcFieldType.Name {
				found = true
				if err := m.mapAny(srcFieldType.Type, srcFieldValue, destFieldType.Type, destFieldValue); err != nil {
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
	if o.mapFunc == nil {
		return nil
	}

	v, err := o.mapFunc(src)
	if err != nil {
		return err
	}
	destValue.Set(reflect.ValueOf(v))
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
	_, found := m.destType.FieldByName(name)
	if !found {
		panic(fmt.Errorf("destination has no field named %s", name))
	}
	opts, found := m.fieldMappings[name]
	if !found {
		opts = Opts{}
	}
	option(&opts)
	m.fieldMappings[name] = opts
	return m
}

func MapSlice[A, B any](slice []A, mapper func(input A) B) (res []B) {
	for _, item := range slice {
		res = append(res, mapper(item))
	}
	return
}
