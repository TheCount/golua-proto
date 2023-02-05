package proto

import (
	rt "github.com/arnodel/golua/runtime"
	"golang.org/x/exp/constraints"
)

// goroutineOverhead is a rough estimate of resources allocated while we have
// to run a goroutine in this package: 2K goroutine stack plus all the stuff
// that is going on.
const goroutineOverhead = 4096

// compliance flags
const (
	cpuIOTimeSafe    = rt.ComplyCpuSafe | rt.ComplyIoSafe | rt.ComplyTimeSafe
	cpuIOMemTimeSafe = cpuIOTimeSafe | rt.ComplyMemSafe
)

// Boolean false and true as Lua values.
var (
	falseValue = rt.BoolValue(false)
	trueValue  = rt.BoolValue(true)
)

// makeClosingVar creates a closing variable which closes the given channel
// and releases the memory resources associated with its function.
func makeClosingVar(toBeClosed chan<- struct{}, releaseMem uint64) rt.Value {
	meta := rt.NewTable()
	meta.Set(rt.StringValue("__close"), rt.FunctionValue(rt.NewGoFunction(
		func(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
			close(toBeClosed)
			t.Runtime.ReleaseMem(releaseMem)
			return c.Next(), nil
		}, "__close", 2, false)))
	ret := rt.NewTable()
	ret.SetMetatable(meta)
	return rt.TableValue(ret)
}

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
	complianceFlags rt.ComplianceFlags,
) {
	goF := rt.NewGoFunction(f, name, nArgs, hasEtc)
	rt.SolemnlyDeclareCompliance(complianceFlags, goF)
	m[name] = rt.FunctionValue(goF)
}

// setTableFunc sets t.name = f.
func setTableFunc(
	t *rt.Table, name string, f rt.GoFunctionFunc, nArgs int, hasEtc bool,
	complianceFlags rt.ComplianceFlags,
) {
	goF := rt.NewGoFunction(f, name, nArgs, hasEtc)
	rt.SolemnlyDeclareCompliance(complianceFlags, goF)
	t.Set(rt.StringValue(name), rt.FunctionValue(goF))
}

// tailMethodCall calls obj:methodName with args and uses the return
// values as argument for the next continuation.
// Should be called as "return tailMethodCall(…)".
func tailMethodCall(
	t *rt.Thread, c rt.Cont, obj rt.Value, methodName string, args []rt.Value,
) (rt.Cont, error) {
	m, err := rt.Index(t, obj, rt.StringValue("Has"))
	if err != nil {
		return nil, err
	}
	return nil, rt.Call(t, m, append([]rt.Value{obj}, args...), c.Next())
}
