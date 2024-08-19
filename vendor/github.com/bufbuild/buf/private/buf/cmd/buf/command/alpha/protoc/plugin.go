// Copyright 2020-2023 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package protoc

import (
	"context"
	"fmt"
	"strings"

	"github.com/bufbuild/buf/private/buf/bufcli"
	"github.com/bufbuild/buf/private/bufpkg/bufimage"
	"github.com/bufbuild/buf/private/bufpkg/bufpluginexec"
	"github.com/bufbuild/buf/private/bufpkg/bufwasm"
	"github.com/bufbuild/buf/private/pkg/app"
	"github.com/bufbuild/buf/private/pkg/command"
	"github.com/bufbuild/buf/private/pkg/storage/storageos"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/pluginpb"
)

type pluginInfo struct {
	// Required
	Out string
	// optional
	Opt []string
	// optional
	Path string
}

func newPluginInfo() *pluginInfo {
	return &pluginInfo{}
}

func executePlugin(
	ctx context.Context,
	logger *zap.Logger,
	storageosProvider storageos.Provider,
	runner command.Runner,
	wasmPluginExecutor bufwasm.PluginExecutor,
	container app.EnvStderrContainer,
	images []bufimage.Image,
	pluginName string,
	pluginInfo *pluginInfo,
) (*pluginpb.CodeGeneratorResponse, error) {
	generator := bufpluginexec.NewGenerator(
		logger,
		storageosProvider,
		runner,
		wasmPluginExecutor,
	)
	var options []bufpluginexec.GenerateOption
	if pluginInfo.Path != "" {
		options = append(options, bufpluginexec.GenerateWithPluginPath(pluginInfo.Path))
	}
	wasmEnabled, err := bufcli.IsAlphaWASMEnabled(container)
	if err != nil {
		return nil, err
	}
	if wasmEnabled {
		options = append(options, bufpluginexec.GenerateWithWASMEnabled())
	}
	response, err := generator.Generate(
		ctx,
		container,
		pluginName,
		bufimage.ImagesToCodeGeneratorRequests(
			images,
			strings.Join(pluginInfo.Opt, ","),
			bufpluginexec.DefaultVersion,
			false,
			false,
		),
		options...,
	)
	if err != nil {
		return nil, fmt.Errorf("--%s_out: %v", pluginName, err)
	}
	return response, nil
}
