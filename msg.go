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

	// msgTableReadOnly is the metatable for protobuf message userdata values
	// that should not be changed.
	msgTableReadOnly *rt.Table

	// msgMethods are the methods for proto messages.
	msgMethods map[string]rt.Value
)

// init initializes msgTable(ReadOnly) and msgMethods.
func init() {
	msgTable = rt.NewTable()
	msgTableReadOnly = rt.NewTable()
	msgMethods = make(map[string]rt.Value)
	setMapFunc(msgMethods, "FullName", msgFullName, 1, false, cpuIOMemTimeSafe)
	setMapFunc(msgMethods, "Has", msgHas, 2, true, cpuIOMemTimeSafe)
	setMapFunc(
		msgMethods, "IsReadOnly", msgIsReadOnly, 1, false, cpuIOMemTimeSafe)
	setMapFunc(msgMethods, "Marshal", msgMarshal, 2, false, cpuIOTimeSafe)
	setMapFunc(msgMethods, "Name", msgName, 1, false, cpuIOMemTimeSafe)
	setMapFunc(msgMethods, "ReadOnly", msgReadOnly, 1, false, cpuIOMemTimeSafe)
	setMapFunc(msgMethods, "Type", msgType, 1, false, cpuIOMemTimeSafe)
	setTableFunc(
		"__eq", msgEqual, 2, false, cpuIOMemTimeSafe, msgTable, msgTableReadOnly)
	setTableFunc("__index", msgIndex, 2, false, cpuIOMemTimeSafe, msgTable)
	setTableFunc(
		"__index", msgIndexReadOnly, 2, false, cpuIOMemTimeSafe, msgTableReadOnly)
	setTableFunc("__newindex", msgNewIndex, 3, false, cpuIOTimeSafe, msgTable)
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

// msgFullName returns the full name of a protobuf message.
func msgFullName(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	ud, _ := c.UserDataArg(0)
	rmsg := ud.Value().(proto.Message).ProtoReflect()
	return pushingString(t, c, string(rmsg.Descriptor().FullName()))
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
	value := protoFieldToLua(rmsg, fd, true)
	return tailMethodCall(t, c, value, "Has", tail)
}

// msgIndex implements the msg[x] operation in Lua.
func msgIndex(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	ud, _ := c.UserDataArg(0)
	msg := ud.Value().(proto.Message)
	k := c.Arg(1)
	if s, ok := k.TryString(); ok {
		return msgIndexString(t, c, msg, s, false)
	}
	if i, ok := k.TryInt(); ok {
		return msgIndexInt(t, c, msg, i, false)
	}
	if ud, ok := k.TryUserData(); ok {
		return msgIndexUserData(t, c, msg, ud, false)
	}
	return c.Next(), nil
}

// msgIndexReadOnly implements the msg[x] operation in Lua.
// Composite fields are returned read-only.
func msgIndexReadOnly(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	ud, _ := c.UserDataArg(0)
	msg := ud.Value().(proto.Message)
	k := c.Arg(1)
	if s, ok := k.TryString(); ok {
		return msgIndexString(t, c, msg, s, true)
	}
	if i, ok := k.TryInt(); ok {
		return msgIndexInt(t, c, msg, i, true)
	}
	if ud, ok := k.TryUserData(); ok {
		return msgIndexUserData(t, c, msg, ud, true)
	}
	return c.Next(), nil
}

// msgIndexInt returns the field with the number indicated by index.
// If readOnly is true, composite fields are returned read-only.
func msgIndexInt(
	t *rt.Thread, c *rt.GoCont, msg proto.Message, idx int64, readOnly bool,
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
	retValue := protoFieldToLua(rmsg, fd, readOnly)
	if retValue.IsNil() {
		return c.Next(), nil
	}
	return c.PushingNext1(t.Runtime, retValue), nil
}

// msgIndexString returns the method of msg named s, or, if it doesn't exist,
// the value of the field named s.
// If readOnly is true, composite fields are returned read-only.
func msgIndexString(
	t *rt.Thread, c *rt.GoCont, msg proto.Message, s string, readOnly bool,
) (rt.Cont, error) {
	if ret, ok := msgMethods[s]; ok {
		return c.PushingNext1(t.Runtime, ret), nil
	}
	rmsg := msg.ProtoReflect()
	fd := rmsg.Descriptor().Fields().ByName(pr.Name(s))
	if fd == nil {
		return c.Next(), nil
	}
	retValue := protoFieldToLua(rmsg, fd, readOnly)
	if retValue.IsNil() {
		return c.Next(), nil
	}
	return c.PushingNext1(t.Runtime, retValue), nil
}

// msgIndexUserData supports indexing msg via various user data types.
// Currently, only pr.FieldDescriptor is supported.
// If readOnly is true, composite fields are returned read-only.
func msgIndexUserData(
	t *rt.Thread, c *rt.GoCont, msg proto.Message, ud *rt.UserData, readOnly bool,
) (rt.Cont, error) {
	rmsg := msg.ProtoReflect()
	switch x := ud.Value().(type) {
	case pr.FieldDescriptor:
		retValue := protoFieldToLua(rmsg, x, readOnly)
		if retValue.IsNil() {
			return c.Next(), nil
		}
		return c.PushingNext1(t.Runtime, retValue), nil
	default:
		return c.Next(), nil
	}
}

// msgName returns the name of a protobuf message.
func msgName(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	ud, _ := c.UserDataArg(0)
	rmsg := ud.Value().(proto.Message).ProtoReflect()
	return pushingString(t, c, string(rmsg.Descriptor().Name()))
}

// msgNewIndex implements the msg[k] = v operation in Lua.
func msgNewIndex(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	ud, _ := c.UserDataArg(0)
	rmsg := ud.Value().(proto.Message).ProtoReflect()
	k := c.Arg(1)
	if s, ok := k.TryString(); ok {
		return msgNewIndexString(t, c, rmsg, s)
	}
	if i, ok := k.TryInt(); ok {
		return msgNewIndexInt(t, c, rmsg, i)
	}
	if ud, ok := k.TryUserData(); ok {
		return msgNewIndexUserData(t, c, rmsg, ud)
	}
	return nil, fmt.Errorf("bad index type %s", k.TypeName())
}

// msgNewIndexFD implements the msg[fd] = v operation in Lua.
func msgNewIndexFD(
	t *rt.Thread, c *rt.GoCont, msg pr.Message, fd pr.FieldDescriptor,
) (rt.Cont, error) {
	if fd.ContainingMessage().FullName() != msg.Descriptor().FullName() {
		return nil, fmt.Errorf(
			"field descriptor '%s' does not belong to message type '%s'",
			fd.FullName(), msg.Descriptor().FullName())
	}
	luaValue := c.Arg(2)
	if luaValue.IsNil() {
		if fd.HasPresence() || fd.IsList() || fd.IsMap() {
			msg.Clear(fd)
			return c.Next(), nil
		}
		return nil, fmt.Errorf("nil value not allowed for field '%s'", fd.Name())
	}
	value, err := luaToProtoValue(fd, luaValue)
	if err != nil {
		return nil, err
	}
	msg.Set(fd, value)
	return c.Next(), nil
}

// msgNewIndexInt implements msg[fieldNumber] = v in Lua.
func msgNewIndexInt(
	t *rt.Thread, c *rt.GoCont, msg pr.Message, fieldNumber int64,
) (rt.Cont, error) {
	if fieldNumber < 0 || fieldNumber > math.MaxInt32 {
		return nil, fmt.Errorf("field number out of bounds: %d", fieldNumber)
	}
	fd := msg.Descriptor().Fields().ByNumber(pr.FieldNumber(fieldNumber))
	if fd == nil {
		return nil, fmt.Errorf("no such field number: %d", fieldNumber)
	}
	return msgNewIndexFD(t, c, msg, fd)
}

// msgNewIndexString implements msg.fieldName = v in Lua.
func msgNewIndexString(
	t *rt.Thread, c *rt.GoCont, msg pr.Message, fieldName string,
) (rt.Cont, error) {
	fd := msg.Descriptor().Fields().ByName(pr.Name(fieldName))
	if fd == nil {
		return nil, fmt.Errorf("no such field: %s", fieldName)
	}
	return msgNewIndexFD(t, c, msg, fd)
}

// msgNewIndexUserData supports setting a message field via various
// user data types.
// Currently, only pr.FieldDescriptor is supported.
func msgNewIndexUserData(
	t *rt.Thread, c *rt.GoCont, msg pr.Message, ud *rt.UserData,
) (rt.Cont, error) {
	switch x := ud.Value().(type) {
	case pr.FieldDescriptor:
		return msgNewIndexFD(t, c, msg, x)
	default:
		return nil, fmt.Errorf("userdata index of type %T not supported", x)
	}
}

// msgIsReadOnly checks whether the message is read-only.
func msgIsReadOnly(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	ud, _ := c.UserDataArg(0)
	return pushingBool(t, c, ud.Metatable() == msgTableReadOnly)
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

// msgReadOnly returns a read-only version of the message.
// If the message is already read-only, this function has no effect.
func msgReadOnly(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	ud, _ := c.UserDataArg(0)
	return pushingUserData(t, c, ud.Value(), msgTableReadOnly)
}

// msgType returns the message type of a protobuf message.
func msgType(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	ud, _ := c.UserDataArg(0)
	msg := ud.Value().(proto.Message)
	return c.PushingNext1(t.Runtime, wrapType(msg.ProtoReflect().Type())), nil
}

// Wrap returns the given protobuf message as a Lua value.
func Wrap(msg proto.Message) rt.Value {
	return wrap(msg, false)
}

// WrapReadOnly returns the given protobuf message as a Lua value.
// The returned message cannot be changed from Lua.
func WrapReadOnly(msg proto.Message) rt.Value {
	return wrap(msg, true)
}

// wrap wraps the given proto message as a Lua value.
// If read-only is true, the wrapped message cannot be changed from Lua.
func wrap(msg proto.Message, readOnly bool) rt.Value {
	if msg == nil {
		return rt.NilValue
	}
	meta := msgTable
	if readOnly {
		meta = msgTableReadOnly
	}
	return rt.UserDataValue(rt.NewUserData(msg, meta))
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
