package gotots

import (
	"encoding"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/abibby/salusa/set"
)

var timeType = reflect.TypeOf(time.Time{})
var jsonMarshalerType = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
var encodingTextMarshalerType = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()

var stringType = reflect.TypeOf("")

var intType = reflect.TypeOf(0)
var int8Type = reflect.TypeOf(int8(0))
var int16Type = reflect.TypeOf(int16(0))
var int32Type = reflect.TypeOf(int32(0))
var int64Type = reflect.TypeOf(int64(0))

var uintType = reflect.TypeOf(uint(0))
var uint8Type = reflect.TypeOf(uint8(0))
var uint16Type = reflect.TypeOf(uint16(0))
var uint32Type = reflect.TypeOf(uint32(0))
var uint64Type = reflect.TypeOf(uint64(0))

var float32Type = reflect.TypeOf(float32(0))
var float64Type = reflect.TypeOf(float64(0))

var boolType = reflect.TypeOf(false)

var basicTypes = map[reflect.Type]string{
	stringType: "string",

	intType:   "number",
	int8Type:  "number",
	int16Type: "number",
	int32Type: "number",
	int64Type: "number",

	uintType:   "number",
	uint8Type:  "number",
	uint16Type: "number",
	uint32Type: "number",
	uint64Type: "number",

	float32Type: "number",
	float64Type: "number",

	boolType: "boolean",
}

func defaultGenerator() *Generator {
	return &Generator{
		TypeMap: map[reflect.Type]string{
			timeType: "string",
		},
	}
}

type Generator struct {
	TypeMap map[reflect.Type]string
}

type Option func(o *Generator) *Generator

func WithType(t reflect.Type, src string) Option {
	return func(o *Generator) *Generator {
		o.TypeMap[t] = src
		return o
	}
}

func GenerateTypes(t reflect.Type, options ...Option) string {
	g := defaultGenerator()
	for _, opt := range options {
		g = opt(g)
	}
	finishedTypes := set.New[reflect.Type]()
	src := ""
	types := []reflect.Type{t}
	allFinished := len(types) == 0
	for !allFinished {
		allFinished = true
		for _, newType := range types {
			if finishedTypes.Has(newType) {
				continue
			}
			allFinished = false

			finishedTypes.Add(newType)
			typeSrc, newTypes := g.generate(newType, true)
			src += "export type " + newType.Name() + " = " + typeSrc + ";\n"
			types = append(types, newTypes...)
		}
	}
	return src
}
func (g *Generator) Generate(t reflect.Type) string {
	src, _ := g.generate(t, false)
	return src
}
func (g *Generator) generate(t reflect.Type, rootType bool) (string, []reflect.Type) {
	src, ok := basicTypes[t]
	if ok {
		return src, []reflect.Type{}
	}

	name := t.Name()
	if !rootType && name != "" {
		return name, []reflect.Type{t}
	}
	src, ok = g.TypeMap[t]
	if ok {
		return src, []reflect.Type{}
	}

	if t.Implements(jsonMarshalerType) {
		fmt.Fprintf(os.Stderr, "%s implements json.Marshaler. The result may not be actuate.\n", t.String())
	} else if t.Implements(encodingTextMarshalerType) {
		fmt.Fprintf(os.Stderr, "%s implements encoding.TextMarshaler. The result may not be actuate.\n", t.String())
	}

	switch t.Kind() {
	case reflect.String:
		return "string", []reflect.Type{}
	case reflect.Bool:
		return "boolean", []reflect.Type{}
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Float32,
		reflect.Float64:
		return "number", []reflect.Type{}
	case reflect.Map:
		return g.generateMap(t)
	case reflect.Struct:
		return g.generateStruct(t)
	case reflect.Slice:
		return g.generateSlice(t)
	case reflect.Pointer:
		return g.generatePointer(t)
	default:
		return "unknown", []reflect.Type{}
	}
}

func (g *Generator) generateStruct(t reflect.Type) (string, []reflect.Type) {
	src := ""
	types := []reflect.Type{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		optional := false
		name, ok := f.Tag.Lookup("json")
		if !ok {
			name = f.Name
		} else {
			parts := strings.Split(name, ",")
			optional = slices.Contains(parts, "omitempty")
			name = parts[0]
		}
		typeSrc, fieldTypes := g.generate(f.Type, false)

		if i > 0 {
			src += "\n"
		}
		optionalStr := ""
		if optional {
			optionalStr = "?"
		}
		src += name + optionalStr + ": " + typeSrc + ";"
		types = append(types, fieldTypes...)
	}
	return "{\n" + tabIn(src) + "\n}", types
}

func (g *Generator) generateSlice(t reflect.Type) (string, []reflect.Type) {
	src, types := g.generate(t.Elem(), false)
	return "(" + src + ")[]", types
}

func (g *Generator) generateMap(t reflect.Type) (string, []reflect.Type) {
	keySrc, keyTypes := g.generate(t.Key(), false)
	valueSrc, valueTypes := g.generate(t.Elem(), false)
	return "Record<" + keySrc + ", " + valueSrc + ">", append(keyTypes, valueTypes...)
}

func (g *Generator) generatePointer(t reflect.Type) (string, []reflect.Type) {
	src, types := g.generate(t.Elem(), false)
	return src + " | null", types
}

func tabIn(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = "    " + line
	}
	return strings.Join(lines, "\n")
}
