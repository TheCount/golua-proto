package proto

import (
	"errors"
	"fmt"

	rt "github.com/arnodel/golua/runtime"
	pr "google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

// protoNew creates a new, empty protobuf message.
func protoNew(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	arg := c.Arg(0)
	if s, ok := arg.TryString(); ok {
		return protoNewString(t, c, s)
	}
	if ud, ok := arg.TryUserData(); ok {
		return protoNewUserData(t, c, ud)
	}
	return nil, fmt.Errorf("invalid argument type %s", arg.TypeName())
}

// protoNewMessageDescriptor creates a new, empty protobuf message from
// the given descriptor. The created message will be dynamic if no
// concrete message type could be found in the registry.
func protoNewMessageDescriptor(
	t *rt.Thread, c *rt.GoCont, md pr.MessageDescriptor,
) (rt.Cont, error) {
	cont, err := protoNewString(t, c, string(md.FullName()))
	if err == nil {
		return cont, nil
	}
	msg := dynamicpb.NewMessage(md)
	return c.PushingNext1(t.Runtime, Wrap(msg)), nil
}

// protoNewMessageType creates a new, empty protobuf message of the given
// type.
func protoNewMessageType(
	t *rt.Thread, c *rt.GoCont, mt pr.MessageType,
) (rt.Cont, error) {
	rmsg := mt.New()
	if rmsg == nil {
		return nil, errors.New("unable to create synthetic message")
	}
	return c.PushingNext1(t.Runtime, Wrap(rmsg.Interface())), nil
}

// protoNewString creates a new, empty protobuf message with fullname given
// by s.
func protoNewString(t *rt.Thread, c *rt.GoCont, s string) (rt.Cont, error) {
	mt, err := protoregistry.GlobalTypes.FindMessageByName(pr.FullName(s))
	if err == nil {
		return protoNewMessageType(t, c, mt)
	}
	mt, err = protoregistry.GlobalTypes.FindMessageByURL(s)
	if err != nil {
		return nil, fmt.Errorf("no such message type: %s", s)
	}
	return protoNewMessageType(t, c, mt)
}

// protoNewUserData creates a new, empty protobuf message based on the
// given user data.
func protoNewUserData(
	t *rt.Thread, c *rt.GoCont, ud *rt.UserData,
) (rt.Cont, error) {
	switch x := ud.Value().(type) {
	case pr.MessageType:
		return protoNewMessageType(t, c, x)
	case pr.MessageDescriptor:
		return protoNewMessageDescriptor(t, c, x)
	default:
		return nil, fmt.Errorf("cannot create message from %T", x)
	}
}
