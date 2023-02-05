package proto

import (
	rt "github.com/arnodel/golua/runtime"
	pr "google.golang.org/protobuf/reflect/protoreflect"
)

var (
	// typeTable is the metatable for protobuf message type userdata values.
	typeTable *rt.Table
)

// init initializes typeTable.
func init() {
	typeTable = rt.NewTable()
	setTableFunc(typeTable, "__eq", typeEqual, 2, false, cpuIOMemTimeSafe)
}

// typeEqual checks two protobuf message types for equality.
func typeEqual(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	lhs, _ := c.UserDataArg(0)
	lhsMT, ok := lhs.Value().(pr.MessageType)
	if !ok {
		return pushingFalse(t, c)
	}
	rhs, _ := c.UserDataArg(1)
	rhsMT, ok := rhs.Value().(pr.MessageType)
	if !ok {
		return pushingFalse(t, c)
	}
	return pushingBool(t, c,
		lhsMT.Descriptor().FullName() == rhsMT.Descriptor().FullName())
}

// wrapType wraps the given message type in a Lua value.
func wrapType(mt pr.MessageType) rt.Value {
	if mt == nil {
		return rt.NilValue
	}
	return rt.UserDataValue(rt.NewUserData(mt, typeTable))
}
