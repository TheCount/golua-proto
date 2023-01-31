package proto

import (
	rt "github.com/arnodel/golua/runtime"
	"golang.org/x/exp/constraints"
)

// Boolean false and true as Lua values.
var (
	falseValue = rt.BoolValue(false)
	trueValue  = rt.BoolValue(true)
)

// pushingBool can be used to return the boolean value b.
func pushingBool(t *rt.Thread, c *rt.GoCont, b bool) (rt.Cont, error) {
	return c.PushingNext1(t.Runtime, rt.BoolValue(b)), nil
}

// pushingFalse can be used to return the value false.
func pushingFalse(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	return c.PushingNext1(t.Runtime, falseValue), nil
}

// pushingTrue can be used to return the value true.
func pushingTrue(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	return c.PushingNext1(t.Runtime, trueValue), nil
}

// pushingInt can be used to return the integer value i.
func pushingInt[T constraints.Integer](
	t *rt.Thread, c *rt.GoCont, i T,
) (rt.Cont, error) {
	return c.PushingNext1(t.Runtime, rt.IntValue(int64(i))), nil
}

// pushingString can be used to return the string value s.
func pushingString(t *rt.Thread, c *rt.GoCont, s string) (rt.Cont, error) {
	return c.PushingNext1(t.Runtime, rt.StringValue(s)), nil
}

// setMapFunc sets m[name] = f.
// Useful if __index does more than just access methods. In this case,
// the methods can be stored in m, and the __index function looks up
// the appropriate method in m.
func setMapFunc(
	m map[string]rt.Value, name string,
	f rt.GoFunctionFunc, nArgs int, hasEtc bool,
) {
	m[name] = rt.FunctionValue(rt.NewGoFunction(f, name, nArgs, hasEtc))
}

// setTableFunc sets t.name = f.
func setTableFunc(
	t *rt.Table, name string, f rt.GoFunctionFunc, nArgs int, hasEtc bool,
) {
	t.Set(rt.StringValue(name),
		rt.FunctionValue(rt.NewGoFunction(f, name, nArgs, hasEtc)))
}
