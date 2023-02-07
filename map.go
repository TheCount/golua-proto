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

	// mapTableReadOnly is the metatable for read-only protobuf map field values.
	mapTableReadOnly *rt.Table

	// mapMethods are the methods for protobuf maps.
	mapMethods map[string]rt.Value
)

// init initializes mapTable(ReadOnly) and mapMethods.
func init() {
	mapTable = rt.NewTable()
	mapTableReadOnly = rt.NewTable()
	mapMethods = make(map[string]rt.Value)
	setMapFunc(mapMethods, "Has", mapHas, 2, true, cpuIOMemTimeSafe)
	setMapFunc(mapMethods,
		"IsReadOnly", mapIsReadOnly, 1, false, cpuIOMemTimeSafe)
	setMapFunc(mapMethods, "Range", mapRange, 1, false, cpuIOMemTimeSafe)
	setMapFunc(mapMethods, "ReadOnly", mapReadOnly, 1, false, cpuIOMemTimeSafe)
	setTableFunc("__index", mapIndex, 2, false, cpuIOMemTimeSafe, mapTable)
	setTableFunc(
		"__index", mapIndexReadOnly, 2, false, cpuIOMemTimeSafe, mapTableReadOnly)
	setTableFunc(
		"__len", mapLen, 1, false, cpuIOMemTimeSafe, mapTable, mapTableReadOnly)
	setTableFunc(
		"__len", mapLen, 1, false, cpuIOMemTimeSafe, mapTable, mapTableReadOnly)
}

// mapHas checks whether the map has the specified key.
func mapHas(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	ud, _ := c.UserDataArg(0)
	mw := ud.Value().(*mapWrapper)
	k := c.Arg(1)
	if s, ok := k.TryString(); ok {
		return mapHasString(t, c, mw, s)
	}
	if i, ok := k.TryInt(); ok {
		return mapHasInt(t, c, mw, i)
	}
	if b, ok := k.TryBool(); ok {
		return mapHasBool(t, c, mw, b)
	}
	return pushingFalse(t, c)
}

// mapHasBool checks whether the map has the specified bool key.
func mapHasBool(
	t *rt.Thread, c *rt.GoCont, mw *mapWrapper, b bool,
) (rt.Cont, error) {
	if mw.field.MapKey().Kind() != pr.BoolKind {
		return pushingFalse(t, c)
	}
	return mapHasTail(t, c, mw, pr.ValueOfBool(b).MapKey())
}

// mapHasInt checks whether the map has the specified integer key.
func mapHasInt(
	t *rt.Thread, c *rt.GoCont, mw *mapWrapper, idx int64,
) (rt.Cont, error) {
	var key pr.MapKey
	switch mw.field.MapKey().Kind() {
	case pr.Int32Kind, pr.Sint32Kind, pr.Sfixed32Kind:
		if idx < math.MinInt16 || idx > math.MaxInt32 {
			return pushingFalse(t, c)
		}
		key = pr.ValueOfInt32(int32(idx)).MapKey()
	case pr.Int64Kind, pr.Sint64Kind, pr.Sfixed64Kind:
		key = pr.ValueOfInt64(idx).MapKey()
	case pr.Uint32Kind, pr.Fixed32Kind:
		if idx < 0 || idx > math.MaxUint32 {
			return pushingFalse(t, c)
		}
		key = pr.ValueOfUint32(uint32(idx)).MapKey()
	case pr.Uint64Kind, pr.Fixed64Kind:
		key = pr.ValueOfUint64(uint64(idx)).MapKey()
	default:
		return pushingFalse(t, c)
	}
	return mapHasTail(t, c, mw, key)
}

// mapHasString checks whether the map has the specified key.
func mapHasString(
	t *rt.Thread, c *rt.GoCont, mw *mapWrapper, s string,
) (rt.Cont, error) {
	if mw.field.MapKey().Kind() != pr.StringKind {
		return pushingFalse(t, c)
	}
	key := pr.ValueOfString(s).MapKey()
	return mapHasTail(t, c, mw, key)
}

// mapHasTail performs the tail call for mapHas if key is present.
func mapHasTail(
	t *rt.Thread, c *rt.GoCont, mw *mapWrapper, key pr.MapKey,
) (rt.Cont, error) {
	if !mw.m.Has(key) {
		return pushingFalse(t, c)
	}
	tail := c.Etc()
	if len(tail) == 0 {
		return pushingTrue(t, c)
	}
	value := protoValueToLua(mw.field, mw.m.Get(key), true)
	return tailMethodCall(t, c, value, "Has", tail)
}

// mapIndex performs the index operation on a map in Lua.
func mapIndex(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	ud, _ := c.UserDataArg(0)
	mw := ud.Value().(*mapWrapper)
	k := c.Arg(1)
	if s, ok := k.TryString(); ok {
		return mapIndexString(t, c, mw, s, false)
	}
	if i, ok := k.TryInt(); ok {
		return mapIndexInt(t, c, mw, i, false)
	}
	if b, ok := k.TryBool(); ok {
		return mapIndexBool(t, c, mw, b, false)
	}
	return c.Next(), nil
}

// mapIndexReadOnly performs the index operation on a map in Lua.
// Composite values are returned read-only.
func mapIndexReadOnly(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	ud, _ := c.UserDataArg(0)
	mw := ud.Value().(*mapWrapper)
	k := c.Arg(1)
	if s, ok := k.TryString(); ok {
		return mapIndexString(t, c, mw, s, true)
	}
	if i, ok := k.TryInt(); ok {
		return mapIndexInt(t, c, mw, i, true)
	}
	if b, ok := k.TryBool(); ok {
		return mapIndexBool(t, c, mw, b, true)
	}
	return c.Next(), nil
}

// mapIndexBool returns the map value at the specified bool key.
// If readOnly is true, composite values are returned read-only.
func mapIndexBool(
	t *rt.Thread, c *rt.GoCont, mw *mapWrapper, b bool, readOnly bool,
) (rt.Cont, error) {
	if mw.field.MapKey().Kind() != pr.BoolKind {
		return c.Next(), nil
	}
	key := pr.ValueOfBool(b).MapKey()
	if !mw.m.Has(key) {
		return c.Next(), nil
	}
	ret := protoValueToLua(mw.field.MapValue(), mw.m.Get(key), readOnly)
	return c.PushingNext1(t.Runtime, ret), nil
}

// mapIndexInt returns the map value at the specified key.
// If readOnly is true, composite values are returned read-only.
func mapIndexInt(
	t *rt.Thread, c *rt.GoCont, mw *mapWrapper, idx int64, readOnly bool,
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
	ret := protoValueToLua(mw.field.MapValue(), mw.m.Get(key), readOnly)
	return c.PushingNext1(t.Runtime, ret), nil
}

// mapIndexString returns the map value at the specified string key.
// If readOnly is true, composite values are returned read-only.
func mapIndexString(
	t *rt.Thread, c *rt.GoCont, mw *mapWrapper, s string, readOnly bool,
) (rt.Cont, error) {
	if ret, ok := mapMethods[s]; ok {
		return c.PushingNext1(t.Runtime, ret), nil
	}
	if mw.field.MapKey().Kind() != pr.StringKind {
		return c.Next(), nil
	}
	key := pr.ValueOfString(s).MapKey()
	if !mw.m.Has(key) {
		return c.Next(), nil
	}
	ret := protoValueToLua(mw.field.MapValue(), mw.m.Get(key), readOnly)
	return c.PushingNext1(t.Runtime, ret), nil
}

// mapIsReadOnly checks whether the list is read-only.
func mapIsReadOnly(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	ud, _ := c.UserDataArg(0)
	return pushingBool(t, c, ud.Metatable() == mapTableReadOnly)
}

// mapLen performs the length operation on a map in Lua.
func mapLen(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	ud, _ := c.UserDataArg(0)
	mw := ud.Value().(*mapWrapper)
	return pushingInt(t, c, mw.m.Len())
}

// mapRange allows ranging over a map.
// The pairs() mechanism cannot be used, see
// https://stackoverflow.com/q/75263097/4838452.
func mapRange(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	type keyValue struct {
		key, value rt.Value
	}
	ud, _ := c.UserDataArg(0)
	mw := ud.Value().(*mapWrapper)
	readOnly := ud.Metatable() == mapTableReadOnly
	state := make(chan keyValue)
	done := make(chan struct{})
	t.Runtime.RequireMem(goroutineOverhead)
	closing := makeClosingVar(done, goroutineOverhead)
	go func() {
		mw.m.Range(func(k pr.MapKey, v pr.Value) bool {
			elt := keyValue{
				key:   protoValueToLua(mw.field.MapKey(), k.Value(), true),
				value: protoValueToLua(mw.field.MapValue(), v, readOnly),
			}
			select {
			case <-done:
				return false
			case state <- elt:
				return true
			}
		})
		elt := keyValue{
			key:   rt.NilValue,
			value: rt.NilValue,
		}
		select {
		case <-done:
		case state <- elt:
		}
	}()
	iteratorFunction := rt.NewGoFunction(
		func(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
			elt := <-state
			return c.PushingNext(t.Runtime, elt.key, elt.value), nil
		}, "iterator", 2, false)
	rt.SolemnlyDeclareCompliance(cpuIOMemTimeSafe, iteratorFunction)
	return c.PushingNext(t.Runtime, rt.FunctionValue(iteratorFunction),
		rt.NilValue, rt.NilValue, closing), nil
}

// mapReadOnly returns a read-only version of the map.
// If the map is already read-only, this function has no effect.
func mapReadOnly(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	ud, _ := c.UserDataArg(0)
	return pushingUserData(t, c, ud.Value(), mapTableReadOnly)
}

// wrapMap wraps the given map from the given map field as a Lua value.
// If read-only is true, the wrapped map cannot be changed from Lua.
func wrapMap(fd pr.FieldDescriptor, m pr.Map, readOnly bool) rt.Value {
	meta := mapTable
	if readOnly {
		meta = mapTableReadOnly
	}
	return rt.UserDataValue(rt.NewUserData(&mapWrapper{
		field: fd,
		m:     m,
	}, meta))
}
