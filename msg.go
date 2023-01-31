package proto

import (
	"errors"
	"fmt"
	"math"

	rt "github.com/arnodel/golua/runtime"
	"google.golang.org/protobuf/proto"
	pr "google.golang.org/protobuf/reflect/protoreflect"
)

var (
	// msgTable is the metatable for protobuf message userdata values.
	msgTable *rt.Table

	// msgMethods are the methods for proto messages.
	msgMethods map[string]rt.Value
)

// init initializes msgTable and msgMethods.
func init() {
	msgTable = rt.NewTable()
	msgMethods = make(map[string]rt.Value)
	setMapFunc(msgMethods, "Has", msgHas, 2, true)
	setMapFunc(msgMethods, "Marshal", msgMarshal, 2, false)
	setMapFunc(msgMethods, "Type", msgType, 1, false)
	setTableFunc(msgTable, "__eq", msgEqual, 2, false)
	setTableFunc(msgTable, "__index", msgIndex, 2, false)
}

// msgEqual checks two protobuf messages for equality in Lua.
func msgEqual(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	lhs, _ := c.UserDataArg(0)
	lhsMsg, ok := lhs.Value().(proto.Message)
	if !ok {
		return pushingFalse(t, c)
	}
	rhs, _ := c.UserDataArg(1)
	rhsMsg, ok := rhs.Value().(proto.Message)
	if !ok {
		return pushingFalse(t, c)
	}
	return pushingBool(t, c, proto.Equal(lhsMsg, rhsMsg))
}

// msgHas checks whether the message has the specified field.
func msgHas(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	ud, _ := c.UserDataArg(0)
	rmsg := ud.Value().(proto.Message).ProtoReflect()
	fieldSpec := c.Arg(1)
	var fd pr.FieldDescriptor
	if fieldName, ok := fieldSpec.TryString(); ok {
		fd = rmsg.Descriptor().Fields().ByName(pr.Name(fieldName))
	} else if fieldNumber, ok := fieldSpec.TryInt(); ok {
		if fieldNumber < 0 || fieldNumber > math.MaxInt32 {
			return pushingFalse(t, c)
		}
		fd = rmsg.Descriptor().Fields().ByNumber(pr.FieldNumber(fieldNumber))
	} else {
		return nil, fmt.Errorf("invalid field spec type '%s'", fieldSpec.TypeName())
	}
	if fd == nil {
		return pushingFalse(t, c)
	}
	if !rmsg.Has(fd) {
		return pushingFalse(t, c)
	}
	tail := c.Etc()
	if len(tail) == 0 {
		return pushingTrue(t, c)
	}
	value := protoFieldToLua(rmsg, fd)
	m, err := rt.Index(t, value, rt.StringValue("Has"))
	if err != nil {
		return nil, fmt.Errorf("%s is not an aggregate", fd.FullName())
	}
	err = rt.Call(t, m, append([]rt.Value{value}, tail...), c.Next())
	return c.Next(), err
}

// msgIndex implements the msg[x] operation in Lua.
func msgIndex(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	ud, _ := c.UserDataArg(0)
	msg := ud.Value().(proto.Message)
	k := c.Arg(1)
	if s, ok := k.TryString(); ok {
		return msgIndexString(t, c, msg, s)
	}
	if i, ok := k.TryInt(); ok {
		return msgIndexInt(t, c, msg, i)
	}
	if ud, ok := k.TryUserData(); ok {
		return msgIndexUserData(t, c, msg, ud)
	}
	return c.Next(), nil
}

// msgIndexInt returns the field with the number indicated by index.
func msgIndexInt(
	t *rt.Thread, c *rt.GoCont, msg proto.Message, idx int64,
) (rt.Cont, error) {
	if idx < math.MinInt32 || idx > math.MaxInt32 {
		return c.Next(), nil
	}
	fieldNumber := pr.FieldNumber(idx)
	rmsg := msg.ProtoReflect()
	fd := rmsg.Descriptor().Fields().ByNumber(fieldNumber)
	if fd == nil {
		return c.Next(), nil
	}
	retValue := protoFieldToLua(rmsg, fd)
	if retValue.IsNil() {
		return c.Next(), nil
	}
	return c.PushingNext1(t.Runtime, retValue), nil
}

// msgIndexString returns the method of msg named s, or, if it doesn't exist,
// the value of the field named s.
func msgIndexString(
	t *rt.Thread, c *rt.GoCont, msg proto.Message, s string,
) (rt.Cont, error) {
	if ret, ok := msgMethods[s]; ok {
		return c.PushingNext1(t.Runtime, ret), nil
	}
	rmsg := msg.ProtoReflect()
	fd := rmsg.Descriptor().Fields().ByName(pr.Name(s))
	if fd == nil {
		return c.Next(), nil
	}
	retValue := protoFieldToLua(rmsg, fd)
	if retValue.IsNil() {
		return c.Next(), nil
	}
	return c.PushingNext1(t.Runtime, retValue), nil
}

// msgIndexUserData supports indexing msg via various user data types.
// Currently, only pr.FieldDescriptor is supported.
func msgIndexUserData(
	t *rt.Thread, c *rt.GoCont, msg proto.Message, ud *rt.UserData,
) (rt.Cont, error) {
	rmsg := msg.ProtoReflect()
	switch x := ud.Value().(type) {
	case pr.FieldDescriptor:
		retValue := protoFieldToLua(rmsg, x)
		if retValue.IsNil() {
			return c.Next(), nil
		}
		return c.PushingNext1(t.Runtime, retValue), nil
	default:
		return c.Next(), nil
	}
}

// msgMarshal marshals a protobuf message to wire-format encoding in Lua.
func msgMarshal(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	ud, _ := c.UserDataArg(0)
	msg := ud.Value().(proto.Message)
	optsValue := c.Arg(1)
	if optsValue.IsNil() {
		return msgMarshalPlain(t, c, msg)
	}
	opts, err := c.TableArg(1)
	if err != nil {
		return nil, err
	}
	return msgMarshalOpts(t, c, msg, opts)
}

// msgMarshalPlain marshals a protobuf message to wire-format encoding in Lua.
func msgMarshalPlain(
	t *rt.Thread, c *rt.GoCont, msg proto.Message,
) (rt.Cont, error) {
	buf, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return pushingString(t, c, string(buf))
}

// msgMarshalOpts marshals a protobuf message to wire-format encoding in Lua
// with the given options.
func msgMarshalOpts(
	t *rt.Thread, c *rt.GoCont, msg proto.Message, opts *rt.Table,
) (rt.Cont, error) {
	// FIXME
	return nil, errors.New("sorry, marshalling with options not supported yet")
}

// msgType returns the message type of a protobuf message.
func msgType(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	ud, _ := c.UserDataArg(0)
	msg := ud.Value().(proto.Message)
	return c.PushingNext1(t.Runtime, wrapType(msg.ProtoReflect().Type())), nil
}

// Wrap returns the given protobuf message as a Lua value.
func Wrap(msg proto.Message) rt.Value {
	if msg == nil {
		return rt.NilValue
	}
	return rt.UserDataValue(rt.NewUserData(msg, msgTable))
}

// Unwrap unwraps the protobuf message from the given lua value.
func Unwrap(luaValue rt.Value) (msg proto.Message, ok bool) {
	ud, ok := luaValue.TryUserData()
	if !ok {
		return
	}
	msg, ok = ud.Value().(proto.Message)
	return
}
