package luasandbox

import (
	"context"
	"strings"
	"testing"

	lua "github.com/yuin/gopher-lua"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestSandboxHasNoIO(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		opts := RunOptions{
			Source: `io.open('service_test.go', 'rb')`,
		}
		if err := newService(&observation.TestContext).Run(context.Background(), opts); err == nil {
			t.Fatalf("expected error")
		} else if !strings.Contains(err.Error(), "attempt to index a non-table object(nil) with key 'open'") {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	t.Run("module", func(t *testing.T) {
		opts := RunOptions{
			Source: `
			local io = require("io")
			io.open('service_test.go', 'rb')
		`,
		}
		if err := newService(&observation.TestContext).Run(context.Background(), opts); err == nil {
			t.Fatalf("expected error")
		} else if !strings.Contains(err.Error(), "module io not found") {
			t.Fatalf("unexpected error: %s", err)
		}
	})
}

func TestRunWithBasicModule(t *testing.T) {
	var stashedValue lua.LValue
	ctx := context.Background()

	api := map[string]lua.LGFunction{
		"add": func(state *lua.LState) int {
			a := state.CheckNumber(1)
			b := state.CheckNumber(2)
			state.Push(a + b)

			return 1
		},
		"stash": func(state *lua.LState) int {
			stashedValue = state.CheckAny(1)
			return 1
		},
	}

	testModule := func(state *lua.LState) int {
		t := state.NewTable()
		state.SetFuncs(t, api)
		state.Push(t)
		return 1
	}

	opts := RunOptions{
		Modules: map[string]lua.LGFunction{
			"testmod": testModule,
		},
		Source: `
			local testmod = require("testmod")
			testmod.stash(testmod.add(3, testmod.add(6, 9)))
		`,
	}
	if err := newService(&observation.TestContext).Run(ctx, opts); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if lua.LVAsNumber(stashedValue) != 18 {
		t.Errorf("unexpected stashed value. want=%d have=%d", 18, stashedValue)
	}
}
