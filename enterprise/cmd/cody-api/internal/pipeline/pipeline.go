package pipeline

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	golua "github.com/yuin/gopher-lua"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-api/internal/pipeline/libs"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-api/internal/pipeline/lua"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox/util"
)

type Pipeline struct {
	name              string
	performCapability libs.CapabilityPerformer
	sandboxService    SandboxService
}

func newPipeline(name string, performCapability libs.CapabilityPerformer) *Pipeline {
	return &Pipeline{
		name:              name,
		performCapability: performCapability,
		sandboxService:    luasandbox.NewService(),
	}
}

func (p *Pipeline) Run(ctx context.Context) (string, error) {
	sandbox, err := p.createSandbox(ctx)
	if err != nil {
		return "", err
	}
	defer sandbox.Close()

	return p.runWithSandbox(ctx, sandbox, luasandbox.RunOptions{
		Timeout: time.Minute,
	})
}

func (p *Pipeline) createSandbox(ctx context.Context) (_ *luasandbox.Sandbox, err error) {
	modules, err := makeModules(p.performCapability)
	if err != nil {
		return nil, err
	}
	luaModules, err := luasandbox.LuaModulesFromFS(lua.Scripts, ".", "sg.cody")
	if err != nil {
		return nil, err
	}
	opts := luasandbox.CreateOptions{
		GoModules:  modules,
		LuaModules: luaModules,
	}
	sandbox, err := p.sandboxService.CreateSandbox(ctx, opts)
	if err != nil {
		return nil, err
	}

	return sandbox, nil
}

func (p *Pipeline) runWithSandbox(ctx context.Context, sandbox *luasandbox.Sandbox, opts luasandbox.RunOptions) (string, error) {
	rawPipeline, err := sandbox.RunScriptNamed(ctx, opts, lua.Scripts, filepath.Join(
		"pipelines",
		fmt.Sprintf("%s.lua", p.name),
	))
	if err != nil {
		return "", err
	}
	tablePipeline, ok := rawPipeline.(*golua.LTable)
	if !ok {
		return "", util.NewTypeError("table", rawPipeline)
	}
	pipeline, err := pipelineFromTable(tablePipeline)
	if err != nil {
		return "", err
	}

	// TODO - capability negotiation
	// fmt.Printf("LOOKING FOR CAPABILITIES %v\n", pipeline.capabilities)

	rawContext, err := sandbox.Call(ctx, opts, pipeline.setup)
	if err != nil {
		return "", err
	}

	// TODO - bulk-fill context

	value, err := sandbox.Call(ctx, opts, pipeline.run, rawContext)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v", value), nil
}

//
//

type luaPipeline struct {
	capabilities []string
	setup        *golua.LFunction
	run          *golua.LFunction
}

func pipelineFromTable(table *golua.LTable) (*luaPipeline, error) {
	pipeline := &luaPipeline{}

	if err := util.DecodeTable(table, map[string]func(golua.LValue) error{
		"capabilities": util.SetStrings(&pipeline.capabilities),
		"setup":        util.SetLuaFunction(&pipeline.setup),
		"run":          util.SetLuaFunction(&pipeline.run),
	}); err != nil {
		return nil, err
	}

	return pipeline, nil
}
