package proto

import (
	"fmt"
	"math"

	rt "github.com/arnodel/golua/runtime"
	"google.golang.org/protobuf/proto"
	pr "google.golang.org/protobuf/reflect/protoreflect"
)

// luaToProtoValue converts the given luaValue to a protobuf value that can
// be assigned to the given field descriptor.
func luaToProtoValue(
	fd pr.FieldDescriptor, luaValue rt.Value,
) (pr.Value, error) {
	switch {
	case fd.IsMap():
		ud, ok := luaValue.TryUserData()
		if !ok {
			// TODO: accept table
			return pr.Value{}, fmt.Errorf("field '%s' expects a map", fd.Name())
		}
		mw, ok := ud.Value().(*mapWrapper)
		switch {
		case !ok:
			return pr.Value{}, fmt.Errorf("field '%s' expects a map", fd.Name())
		case fd.MapKey().Kind() != mw.field.MapKey().Kind():
			return pr.Value{}, fmt.Errorf(
				"field '%s' expects map with %s keys, got %s",
				fd.Name(), fd.MapKey().Kind(), mw.field.MapKey().Kind())
		case fd.MapValue().Kind() != mw.field.MapValue().Kind():
			return pr.Value{}, fmt.Errorf(
				"field '%s' expects map with %s values, got %s",
				fd.Name(), fd.MapValue().Kind(), mw.field.MapValue().Kind())
		case fd.MapValue().Kind() == pr.EnumKind:
			lhsED, rhsED := fd.MapValue().Enum(), mw.field.MapValue().Enum()
			if lhsED.FullName() != rhsED.FullName() {
				return pr.Value{}, fmt.Errorf(
					"field '%s' expects enum %s values, got %s",
					fd.Name(), lhsED.Name(), rhsED.Name())
			}
		case fd.MapValue().Kind() == pr.MessageKind:
			lhsMD, rhsMD := fd.MapValue().Message(), mw.field.MapValue().Message()
			if lhsMD.FullName() != rhsMD.FullName() {
				return pr.Value{}, fmt.Errorf(
					"field '%s' expects message %s values, got %s",
					fd.Name(), lhsMD.Name(), rhsMD.Name())
			}
		}
		return pr.ValueOfMap(mw.m), nil
	case fd.IsList():
		ud, ok := luaValue.TryUserData()
		if !ok {
			// TODO: accept table
			return pr.Value{}, fmt.Errorf("field '%s' expects a list", fd.Name())
		}
		lw, ok := ud.Value().(*listWrapper)
		switch {
		case !ok:
			return pr.Value{}, fmt.Errorf("field '%s' expects a list", fd.Name())
		case fd.Kind() != lw.field.Kind():
			return pr.Value{}, fmt.Errorf(
				"field '%s' expects list with %s values, got %s",
				fd.Name(), fd.Kind(), lw.field.Kind())
		case fd.Kind() == pr.EnumKind:
			lhsED, rhsED := fd.Enum(), lw.field.Enum()
			if lhsED.FullName() != rhsED.FullName() {
				return pr.Value{}, fmt.Errorf(
					"field '%s' expects enum %s values, got %s",
					fd.Name(), lhsED.Name(), rhsED.Name())
			}
		case fd.Kind() == pr.MessageKind:
			lhsMD, rhsMD := fd.Message(), lw.field.Message()
			if lhsMD.FullName() != rhsMD.FullName() {
				return pr.Value{}, fmt.Errorf(
					"field '%s' expects message %s values, got %s",
					fd.Name(), lhsMD.Name(), rhsMD.Name())
			}
		}
		return pr.ValueOfList(lw.list), nil
	case fd.Kind() == pr.BoolKind:
		b, ok := luaValue.TryBool()
		if !ok {
			return pr.Value{}, fmt.Errorf("expected bool, got %s",
				luaValue.TypeName())
		}
		return pr.ValueOfBool(b), nil
	case fd.Kind() == pr.EnumKind:
		var n pr.EnumNumber
		if i, ok := luaValue.TryInt(); ok {
			if i < math.MinInt32 || i > math.MaxInt32 {
				return pr.Value{}, fmt.Errorf("enum number %d out of bounds", i)
			}
			n = pr.EnumNumber(i)
		} else if s, ok := luaValue.TryString(); ok {
			evd := fd.Enum().Values().ByName(pr.Name(s))
			if evd == nil {
				return pr.Value{}, fmt.Errorf("enum value '%s' not found", s)
			}
			n = evd.Number()
		} else {
			return pr.Value{}, fmt.Errorf("expected enum, got %s",
				luaValue.TypeName())
		}
		return pr.ValueOfEnum(n), nil
	case fd.Kind() == pr.Int32Kind, fd.Kind() == pr.Sint32Kind,
		fd.Kind() == pr.Sfixed32Kind:
		i, ok := luaValue.TryInt()
		if !ok {
			return pr.Value{}, fmt.Errorf("expected integer, got %s",
				luaValue.TypeName())
		}
		if i < math.MinInt32 || i > math.MaxInt32 {
			return pr.Value{}, fmt.Errorf("%s field out of bounds: %d", fd.Kind(), i)
		}
		return pr.ValueOfInt32(int32(i)), nil
	case fd.Kind() == pr.Uint32Kind, fd.Kind() == pr.Fixed32Kind:
		i, ok := luaValue.TryInt()
		if !ok {
			return pr.Value{}, fmt.Errorf("expected integer, got %s",
				luaValue.TypeName())
		}
		if i < 0 || i > math.MaxUint32 {
			return pr.Value{}, fmt.Errorf("%s field out of bounds: %d", fd.Kind(), i)
		}
		return pr.ValueOfUint32(uint32(i)), nil
	case fd.Kind() == pr.Int64Kind, fd.Kind() == pr.Sint64Kind,
		fd.Kind() == pr.Sfixed64Kind:
		i, ok := luaValue.TryInt()
		if !ok {
			return pr.Value{}, fmt.Errorf("expected integer, got %s",
				luaValue.TypeName())
		}
		return pr.ValueOfInt64(i), nil
	case fd.Kind() == pr.Uint64Kind, fd.Kind() == pr.Fixed64Kind:
		i, ok := luaValue.TryInt()
		if !ok {
			return pr.Value{}, fmt.Errorf("expected integer, got %s",
				luaValue.TypeName())
		}
		return pr.ValueOfUint64(uint64(i)), nil
	case fd.Kind() == pr.FloatKind:
		var f float32
		if i, ok := luaValue.TryInt(); ok {
			f = float32(i)
		} else if f64, ok := luaValue.TryFloat(); ok {
			f = float32(f64)
		} else {
			return pr.Value{}, fmt.Errorf("expected number, got %s",
				luaValue.TypeName())
		}
		return pr.ValueOfFloat32(f), nil
	case fd.Kind() == pr.DoubleKind:
		var f float64
		if i, ok := luaValue.TryInt(); ok {
			f = float64(i)
		} else if f64, ok := luaValue.TryFloat(); ok {
			f = f64
		} else {
			return pr.Value{}, fmt.Errorf("expected number, got %s",
				luaValue.TypeName())
		}
		return pr.ValueOfFloat64(f), nil
	case fd.Kind() == pr.StringKind:
		s, ok := luaValue.TryString()
		if !ok {
			return pr.Value{}, fmt.Errorf("expected string, got %s",
				luaValue.TypeName())
		}
		return pr.ValueOfString(s), nil
	case fd.Kind() == pr.BytesKind:
		s, ok := luaValue.TryString()
		if !ok {
			return pr.Value{}, fmt.Errorf("expected string, got %s",
				luaValue.TypeName())
		}
		return pr.ValueOfBytes([]byte(s)), nil
	case fd.Kind() == pr.MessageKind:
		ud, ok := luaValue.TryUserData()
		if !ok {
			return pr.Value{}, fmt.Errorf("expected userdata, got %s",
				luaValue.TypeName())
		}
		msg, ok := ud.Value().(proto.Message)
		if !ok {
			return pr.Value{}, fmt.Errorf("expected message, got %T", ud.Value())
		}
		rmsg := msg.ProtoReflect()
		lhsMD, rhsMD := fd.Message(), rmsg.Descriptor()
		if lhsMD.FullName() != rhsMD.FullName() {
			return pr.Value{}, fmt.Errorf(
				"field '%s' expects message %s values, got %s",
				fd.Name(), lhsMD.Name(), rhsMD.Name())
		}
		return pr.ValueOfMessage(rmsg), nil
	default:
		return pr.Value{}, fmt.Errorf("unsupported kind %s", fd.Kind())
	}
}

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
