package proto

import (
	"github.com/arnodel/golua/lib/packagelib"
	rt "github.com/arnodel/golua/runtime"
)

// LibLoader can load the proto package.
var LibLoader = packagelib.Loader{
	Load: load,
	Name: "proto",
}

// load builds the proto package and returns it.
func load(r *rt.Runtime) (rt.Value, func()) {
	pkg := rt.NewTable()
	rt.SolemnlyDeclareCompliance(
		rt.ComplyCpuSafe|rt.ComplyIoSafe|rt.ComplyTimeSafe,
		r.SetEnvGoFunc(pkg, "new", protoNew, 1, false),
	)
	return rt.TableValue(pkg), func() {}
}
