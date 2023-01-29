package proto

import (
	rt "github.com/arnodel/golua/runtime"
	pr "google.golang.org/protobuf/reflect/protoreflect"
)

// protoFieldToLua converts the given protobuf value specified as a message
// and a message field to a Lua value.
// If the value is not valid, nil is returned.
func protoFieldToLua(rmsg pr.Message, fd pr.FieldDescriptor) rt.Value {
	if fd.HasPresence() && !rmsg.Has(fd) {
		return rt.NilValue
	}
	return protoValueToLua(fd, rmsg.Get(fd))
}

// protoValueToLua returns the given protobuf value from the given field
// as a Lua value.
func protoValueToLua(fd pr.FieldDescriptor, value pr.Value) rt.Value {
	switch x := value.Interface().(type) {
	case bool:
		return rt.BoolValue(x)
	case int32:
		return rt.IntValue(int64(x))
	case int64:
		return rt.IntValue(x)
	case uint32:
		return rt.IntValue(int64(x))
	case uint64:
		return rt.IntValue(int64(x))
	case float32:
		return rt.FloatValue(float64(x))
	case float64:
		return rt.FloatValue(x)
	case string:
		return rt.StringValue(x)
	case []byte:
		return rt.StringValue(string(x))
	case pr.EnumNumber:
		return rt.IntValue(int64(x))
	case pr.Message:
		return Wrap(x.Interface())
	case pr.List:
		return wrapList(fd, x)
	case pr.Map:
		return wrapMap(fd, x)
	default:
		return rt.NilValue
	}
}
