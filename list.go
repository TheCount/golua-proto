package proto

import (
	rt "github.com/arnodel/golua/runtime"
	pr "google.golang.org/protobuf/reflect/protoreflect"
)

// listWrapper wraps a protobuf repeated field value for Lua.
type listWrapper struct {
	// field is the field accepting list.
	field pr.FieldDescriptor

	// list is the wrapped list.
	list pr.List
}

var (
	// listTable is the metatable for protobuf repeated field values.
	listTable *rt.Table

	// listMethods are the methods for protobuf lists.
	listMethods map[string]rt.Value
)

// init initializes listTable and listMethods.
func init() {
	listTable = rt.NewTable()
	listMethods = make(map[string]rt.Value)
	setMapFunc(listMethods, "Range", listRange, 1, false, cpuIOMemTimeSafe)
	setTableFunc(listTable, "__index", listIndex, 2, false, cpuIOMemTimeSafe)
	setTableFunc(listTable, "__len", listLen, 1, false, cpuIOMemTimeSafe)
	setTableFunc(listTable, "__pairs", listPairs, 1, false, cpuIOMemTimeSafe)
}

// listIndex performs the index operation on a list in Lua.
func listIndex(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	ud, _ := c.UserDataArg(0)
	lw := ud.Value().(*listWrapper)
	k := c.Arg(1)
	if i, ok := k.TryInt(); ok {
		return listIndexInt(t, c, lw, i)
	}
	if s, ok := k.TryString(); ok {
		return listIndexString(t, c, lw, s)
	}
	return c.Next(), nil
}

// listIndexInt returns the list item at the specified index.
func listIndexInt(
	t *rt.Thread, c *rt.GoCont, lw *listWrapper, idx int64,
) (rt.Cont, error) {
	if idx <= 0 || idx > int64(lw.list.Len()) {
		return c.Next(), nil
	}
	ret := protoValueToLua(lw.field, lw.list.Get(int(idx-1)))
	return c.PushingNext1(t.Runtime, ret), nil
}

// listIndexString returns the method named s.
func listIndexString(
	t *rt.Thread, c *rt.GoCont, lw *listWrapper, s string,
) (rt.Cont, error) {
	if ret, ok := listMethods[s]; ok {
		return c.PushingNext1(t.Runtime, ret), nil
	}
	return c.Next(), nil
}

// listLen performs the length operation on a list in Lua.
func listLen(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	ud, _ := c.UserDataArg(0)
	lw := ud.Value().(*listWrapper)
	return pushingInt(t, c, lw.list.Len())
}

// listPairs implements the __pairs mechanism for ranging over the list.
// See https://www.lua.org/manual/5.4/manual.html#6.1.
func listPairs(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	ud, _ := c.UserDataArg(0)
	lw := ud.Value().(*listWrapper)
	iteratorFunction := rt.NewGoFunction(
		func(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
			if lw.list.Len() == 0 {
				return c.PushingNext(t.Runtime, rt.NilValue, rt.NilValue), nil
			}
			ctrl := c.Arg(1)
			if ctrl.IsNil() {
				return c.PushingNext(t.Runtime, rt.IntValue(1),
					protoValueToLua(lw.field, lw.list.Get(0))), nil
			}
			idx := int(ctrl.AsInt())
			if idx == lw.list.Len() {
				return c.PushingNext(t.Runtime, rt.NilValue, rt.NilValue), nil
			}
			return c.PushingNext(t.Runtime, rt.IntValue(int64(idx)+1),
				protoValueToLua(lw.field, lw.list.Get(idx))), nil
		}, "iterator", 2, false)
	rt.SolemnlyDeclareCompliance(cpuIOMemTimeSafe, iteratorFunction)
	return c.PushingNext(t.Runtime, rt.FunctionValue(iteratorFunction),
		rt.NilValue, rt.NilValue), nil
}

// listRange allows ranging over a list.
// While the pairs() mechanism is more idiomatic,
// it is not available for maps, so we provide this analogous
// mechanism for lists as well.
func listRange(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	err, _ := rt.Metacall(t, c.Arg(0), "__pairs", nil, c.Next())
	return nil, err
}

// wrapList wraps the given list from the given list field as a Lua value.
func wrapList(fd pr.FieldDescriptor, list pr.List) rt.Value {
	return rt.UserDataValue(rt.NewUserData(&listWrapper{
		field: fd,
		list:  list,
	}, listTable))
}
