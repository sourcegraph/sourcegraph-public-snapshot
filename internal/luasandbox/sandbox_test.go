package luasandbox

import (
	"context"
	"strings"
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestSandboxHasNoIO(t *testing.T) {
	ctx := context.Background()

	sandbox, err := newService(&observation.TestContext).CreateSandbox(ctx, CreateOptions{})
	if err != nil {
		t.Fatalf("unexpected error creating sandbox: %s", err)
	}
	defer sandbox.Close()

	t.Run("default", func(t *testing.T) {
		script := `
			io.open('service_test.go', 'rb')
		`
		if _, err := sandbox.RunScript(ctx, RunOptions{}, script); err == nil {
			t.Fatalf("expected error running script")
		} else if !strings.Contains(err.Error(), "attempt to index a non-table object(nil) with key 'open'") {
			t.Fatalf("unexpected error running script: %s", err)
		}
	})

	t.Run("module", func(t *testing.T) {
		script := `
			local io = require("io")
			io.open('service_test.go', 'rb')
		`
		if _, err := sandbox.RunScript(ctx, RunOptions{}, script); err == nil {
			t.Fatalf("expected error running script")
		} else if !strings.Contains(err.Error(), "module io not found") {
			t.Fatalf("unexpected error running script: %s", err)
		}
	})
}

func TestSandboxMaxTimeout(t *testing.T) {
	ctx := context.Background()

	sandbox, err := newService(&observation.TestContext).CreateSandbox(ctx, CreateOptions{})
	if err != nil {
		t.Fatalf("unexpected error creating sandbox: %s", err)
	}
	defer sandbox.Close()

	script := `
		while true do end
	`
	if _, err := sandbox.RunScript(ctx, RunOptions{Timeout: time.Millisecond}, script); err == nil {
		t.Fatalf("expected error running script")
	} else if !strings.Contains(err.Error(), context.DeadlineExceeded.Error()) {
		t.Fatalf("unexpected error running script: %#v", err)
	}
}

func TestRunScript(t *testing.T) {
	ctx := context.Background()

	sandbox, err := newService(&observation.TestContext).CreateSandbox(ctx, CreateOptions{})
	if err != nil {
		t.Fatalf("unexpected error creating sandbox: %s", err)
	}
	defer sandbox.Close()

	script := `
		return 38 + 4
	`
	retValue, err := sandbox.RunScript(ctx, RunOptions{}, script)
	if err != nil {
		t.Fatalf("unexpected error running script: %s", err)
	}
	if lua.LVAsNumber(retValue) != 42 {
		t.Errorf("unexpected return value. want=%d have=%v", 42, retValue)
	}
}

func TestModule(t *testing.T) {
	var stashedValue lua.LValue

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

	ctx := context.Background()

	sandbox, err := newService(&observation.TestContext).CreateSandbox(ctx, CreateOptions{
		Modules: map[string]lua.LGFunction{
			"testmod": testModule,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error creating sandbox: %s", err)
	}
	defer sandbox.Close()

	script := `
		local testmod = require("testmod")
		testmod.stash(testmod.add(3, testmod.add(6, 9)))
		return testmod.add(38, 4)
	`
	retValue, err := sandbox.RunScript(ctx, RunOptions{}, script)
	if err != nil {
		t.Fatalf("unexpected error running script: %s", err)
	}
	if lua.LVAsNumber(retValue) != 42 {
		t.Errorf("unexpected return value. want=%d have=%v", 42, retValue)
	}
	if lua.LVAsNumber(stashedValue) != 18 {
		t.Errorf("unexpected stashed value. want=%d have=%d", 18, stashedValue)
	}
}

func TestCall(t *testing.T) {
	ctx := context.Background()

	sandbox, err := newService(&observation.TestContext).CreateSandbox(ctx, CreateOptions{})
	if err != nil {
		t.Fatalf("unexpected error creating sandbox: %s", err)
	}
	defer sandbox.Close()

	script := `
		local value = 0
		local callback = function(multiplier)
			value = value + 1
			return value * multiplier
		end
		return callback
	`
	retValue, err := sandbox.RunScript(ctx, RunOptions{}, script)
	if err != nil {
		t.Fatalf("unexpected error running script: %s", err)
	}
	callback, ok := retValue.(*lua.LFunction)
	if !ok {
		t.Fatalf("unexpected return type")
	}

	multiplier := 6
	for value := 1; value < 5; value++ {
		expectedValue := value * multiplier

		if retValue, err := sandbox.Call(ctx, RunOptions{}, callback, multiplier); err != nil {
			t.Fatalf("unexpected error invoking callback: %s", err)
		} else if int(lua.LVAsNumber(retValue)) != expectedValue {
			t.Errorf("unexpected value from callback #%d. want=%d have=%v", value, expectedValue, retValue)
		}
	}
}
