package luasandbox

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	lua "github.com/yuin/gopher-lua"

	"github.com/sourcegraph/sourcegraph/internal/luasandbox/util"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestSandboxHasNoIO(t *testing.T) {
	ctx := context.Background()

	sandbox, err := newService(observation.TestContextTB(t)).CreateSandbox(ctx, CreateOptions{})
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

func TestSandboxHasFun(t *testing.T) {
	ctx := context.Background()

	sandbox, err := newService(observation.TestContextTB(t)).CreateSandbox(ctx, CreateOptions{})
	if err != nil {
		t.Fatalf("unexpected error creating sandbox: %s", err)
	}
	defer sandbox.Close()

	script := `
		return require("fun")
	`

	if val, err := sandbox.RunScript(ctx, RunOptions{}, script); err == nil {
		if val.Type() != lua.LTTable {
			t.Fatalf("expected 'fun' to be a table: %s", val)
		}
	} else if strings.Contains(err.Error(), "module fun not found") {
		t.Fatalf("unexpected error running script: %s", err)
	}
}

func TestSandboxMaxTimeout(t *testing.T) {
	ctx := context.Background()

	sandbox, err := newService(observation.TestContextTB(t)).CreateSandbox(ctx, CreateOptions{})
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

	sandbox, err := newService(observation.TestContextTB(t)).CreateSandbox(ctx, CreateOptions{})
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
		"add": util.WrapLuaFunction(func(state *lua.LState) error {
			state.Push(state.CheckNumber(1) + state.CheckNumber(2))
			return nil
		}),
		"stash": util.WrapLuaFunction(func(state *lua.LState) error {
			stashedValue = state.CheckAny(1)
			return nil
		}),
	}

	ctx := context.Background()

	sandbox, err := newService(observation.TestContextTB(t)).CreateSandbox(ctx, CreateOptions{
		GoModules: map[string]lua.LGFunction{
			"testmod": util.CreateModule(api),
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

	sandbox, err := newService(observation.TestContextTB(t)).CreateSandbox(ctx, CreateOptions{})
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

func TestCallGenerator(t *testing.T) {
	ctx := context.Background()

	sandbox, err := newService(observation.TestContextTB(t)).CreateSandbox(ctx, CreateOptions{})
	if err != nil {
		t.Fatalf("unexpected error creating sandbox: %s", err)
	}
	defer sandbox.Close()

	script := `
		local value = 0
		local callback = function(upperBound, multiplier)
			for i=1,upperBound-1 do
				value = value + 1
				coroutine.yield(value * multiplier)
			end

			return (value + 1) * multiplier
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

	upperBound := 5
	multiplier := 6

	retValues, err := sandbox.CallGenerator(ctx, RunOptions{}, callback, upperBound, multiplier)
	if err != nil {
		t.Fatalf("unexpected error invoking callback: %s", err)
	}

	values := make([]int, 0, len(retValues))
	for _, retValue := range retValues {
		values = append(values, int(lua.LVAsNumber(retValue)))
	}
	expectedValues := []int{
		6,  // 1 * 6
		12, // 2*6
		18, // 3*6
		24, // 4*6
		30, //  5 * 6 (the return)
	}
	if diff := cmp.Diff(expectedValues, values); diff != "" {
		t.Errorf("unexpected file contents (-want +got):\n%s", diff)
	}
}

func TestDefaultLuaModulesFilesLoad(t *testing.T) {
	ctx := context.Background()

	modules, err := DefaultGoModules.Init()
	if err != nil {
		t.Fatalf("unexpected error loading modules: %s", err)
	}
	sandbox, err := newService(observation.TestContextTB(t)).CreateSandbox(ctx, CreateOptions{
		GoModules: modules,
	})
	if err != nil {
		t.Fatalf("unexpected error creating sandbox: %s", err)
	}
	defer sandbox.Close()

	for mod, script := range DefaultLuaModules {
		_, err := sandbox.RunScript(ctx, RunOptions{}, script)
		if err != nil {
			t.Fatalf("unexpected error loading runtime file %s: %s", mod, err)
		}
	}
}
