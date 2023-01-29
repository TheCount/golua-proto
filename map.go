package proto

import (
	"math"

	rt "github.com/arnodel/golua/runtime"
	pr "google.golang.org/protobuf/reflect/protoreflect"
)

// mapWrapper wraps a protobuf map field value for Lua.
type mapWrapper struct {
	// field is the field accepting m.
	field pr.FieldDescriptor

	// m is the wrapped map.
	m pr.Map
}

var (
	// mapTable is the metatable for protobuf map field values.
	mapTable *rt.Table
)

// init initializes mapTable.
func init() {
	mapTable = rt.NewTable()
	setTableFunc(mapTable, "__index", mapIndex, 2, false)
	setTableFunc(mapTable, "__len", mapLen, 1, false)
}

// mapIndex performs the index operation on a map in Lua.
func mapIndex(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	ud, _ := c.UserDataArg(0)
	mw := ud.Value().(*mapWrapper)
	k := c.Arg(1)
	if s, ok := k.TryString(); ok {
		return mapIndexString(t, c, mw, s)
	}
	if i, ok := k.TryInt(); ok {
		return mapIndexInt(t, c, mw, i)
	}
	if b, ok := k.TryBool(); ok {
		return mapIndexBool(t, c, mw, b)
	}
	return c.Next(), nil
}

// mapIndexBool returns the map value at the specified bool key.
func mapIndexBool(
	t *rt.Thread, c *rt.GoCont, mw *mapWrapper, b bool,
) (rt.Cont, error) {
	if mw.field.MapKey().Kind() != pr.BoolKind {
		return c.Next(), nil
	}
	key := pr.ValueOfBool(b).MapKey()
	if !mw.m.Has(key) {
		return c.Next(), nil
	}
	ret := protoValueToLua(mw.field.MapValue(), mw.m.Get(key))
	return c.PushingNext1(t.Runtime, ret), nil
}

// mapIndexInt returns the map value at the specified key.
func mapIndexInt(
	t *rt.Thread, c *rt.GoCont, mw *mapWrapper, idx int64,
) (rt.Cont, error) {
	var key pr.MapKey
	switch mw.field.MapKey().Kind() {
	case pr.Int32Kind, pr.Sint32Kind, pr.Sfixed32Kind:
		if idx < math.MinInt16 || idx > math.MaxInt32 {
			return c.Next(), nil
		}
		key = pr.ValueOfInt32(int32(idx)).MapKey()
	case pr.Int64Kind, pr.Sint64Kind, pr.Sfixed64Kind:
		key = pr.ValueOfInt64(idx).MapKey()
	case pr.Uint32Kind, pr.Fixed32Kind:
		if idx < 0 || idx > math.MaxUint32 {
			return c.Next(), nil
		}
		key = pr.ValueOfUint32(uint32(idx)).MapKey()
	case pr.Uint64Kind, pr.Fixed64Kind:
		key = pr.ValueOfUint64(uint64(idx)).MapKey()
	default:
		return c.Next(), nil
	}
	if !mw.m.Has(key) {
		return c.Next(), nil
	}
	ret := protoValueToLua(mw.field.MapValue(), mw.m.Get(key))
	return c.PushingNext1(t.Runtime, ret), nil
}

// mapIndexString returns the map value at the specified string key.
func mapIndexString(
	t *rt.Thread, c *rt.GoCont, mw *mapWrapper, s string,
) (rt.Cont, error) {
	if mw.field.MapKey().Kind() != pr.StringKind {
		return c.Next(), nil
	}
	key := pr.ValueOfString(s).MapKey()
	if !mw.m.Has(key) {
		return c.Next(), nil
	}
	ret := protoValueToLua(mw.field.MapValue(), mw.m.Get(key))
	return c.PushingNext1(t.Runtime, ret), nil
}

// mapLen performs the length operation on a map in Lua.
func mapLen(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	ud, _ := c.UserDataArg(0)
	mw := ud.Value().(*mapWrapper)
	return pushingInt(t, c, mw.m.Len())
}

// wrapMap wraps the given map from the given map field as a Lua value.
func wrapMap(fd pr.FieldDescriptor, m pr.Map) rt.Value {
	return rt.UserDataValue(rt.NewUserData(&mapWrapper{
		field: fd,
		m:     m,
	}, mapTable))
}
