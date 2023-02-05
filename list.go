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
)

// init initializes listTable.
func init() {
	listTable = rt.NewTable()
	setTableFunc(listTable, "__index", listIndex, 2, false, cpuIOMemTimeSafe)
	setTableFunc(listTable, "__len", listLen, 1, false, cpuIOMemTimeSafe)
}

// listIndex performs the index operation on a list in Lua.
func listIndex(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	ud, _ := c.UserDataArg(0)
	lw := ud.Value().(*listWrapper)
	k := c.Arg(1)
	if i, ok := k.TryInt(); ok {
		return listIndexInt(t, c, lw, i)
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

// listLen performs the length operation on a list in Lua.
func listLen(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	ud, _ := c.UserDataArg(0)
	lw := ud.Value().(*listWrapper)
	return pushingInt(t, c, lw.list.Len())
}

// wrapList wraps the given list from the given list field as a Lua value.
func wrapList(fd pr.FieldDescriptor, list pr.List) rt.Value {
	return rt.UserDataValue(rt.NewUserData(&listWrapper{
		field: fd,
		list:  list,
	}, listTable))
}
