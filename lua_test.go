package proto_test

import (
	"testing"

	proto "github.com/TheCount/golua-proto"
	"github.com/arnodel/golua/lib"
	"github.com/arnodel/golua/lib/base"
	"github.com/arnodel/golua/lib/packagelib"
	"github.com/arnodel/golua/luatesting"
	rt "github.com/arnodel/golua/runtime"
	_ "google.golang.org/protobuf/types/known/durationpb"
	_ "google.golang.org/protobuf/types/known/timestamppb"
)

// TestProtoLib runs all Lua tests in the proto library.
func TestProtoLib(t *testing.T) {
	luatesting.RunLuaTestsInDir(t, "lua", func(r *rt.Runtime) func() {
		return lib.LoadLibs(r,
			base.LibLoader, packagelib.LibLoader, proto.LibLoader)
	})
}
