import URI from "vs/base/common/uri";
import { IEnvironmentService } from "vs/platform/environment/common/environment";
import { IExtensionService } from "vs/platform/extensions/common/extensions";
import { IWorkspaceContextService, WorkspaceContextService } from "vs/platform/workspace/common/workspace";
import { IExtensionApiFactory, createApiFactory } from "vs/workbench/api/node/extHost.api.impl";
import { ExtHostContext, IInitData, MainContext } from "vs/workbench/api/node/extHost.protocol";
import { IWorkspaceConfigurationService, IWorkspaceConfigurationValues } from "vs/workbench/services/configuration/common/configuration";
import { IThreadService } from "vs/workbench/services/thread/common/threadService";

import { ExtHostThreadService } from "sourcegraph/ext/extHostThreadService";
import { InitializationOptions } from "sourcegraph/ext/protocol";
import { bulkEnable as bulkEnableFeatures } from "sourcegraph/util/features";
import { EnvironmentService } from "sourcegraph/workbench/environmentService";
import { ExtensionService } from "sourcegraph/workbench/extensionService";
import "sourcegraph/workbench/overrides/package";

const initOpts: InitializationOptions = JSON.parse(decodeURIComponent(self.location.hash.slice(1)));
bulkEnableFeatures(initOpts.features);

// TODO(sqs): pass through the zap ref
self["__tmpZapRef"] = initOpts.tmpZapRef;

/**
 * createExtensionAPI returns an extension API factory, which creates
 * a handle for the vscode.d.ts extension API.
 */
export function createExtensionAPI(
	extHostThreadService: IThreadService = new ExtHostThreadService(),
	environmentService: IEnvironmentService = new EnvironmentService() as IEnvironmentService,
	contextService: IWorkspaceContextService = new WorkspaceContextService({
		// Get the workspace from the URL fragment, which was added to
		// our worker URL in sourcegraph/ext/main.
		resource: URI.parse(initOpts.workspace),
	}),
	extensionService: IExtensionService = new ExtensionService(),
	configurationService: IWorkspaceConfigurationService = {
		// TODO(sqs): Use the real configuration service.
		values(): IWorkspaceConfigurationValues {
			const value = (v: any) => ({
				workspace: v,
				value: v,
				default: v,
				user: v,
			});
			return {
				zap: value({ enable: true, share: { selections: true }, overwrite: false }),
			};
		},
	} as any,
): IExtensionApiFactory {
	// HACK: Deleting these is necessary. If we don't do it, then
	// createApiFactory eventually tries to reference them, and other
	// things related to them.
	delete MainContext.MainThreadTerminalService;
	delete MainContext.MainProcessExtensionService;
	delete ExtHostContext.ExtHostTerminalService;

	const initData: IInitData = {
		parentPid: 0,
		environment: {
			appSettingsHome: environmentService.appSettingsHome,
			disableExtensions: environmentService.disableExtensions,
			userExtensionsHome: environmentService.extensionsPath,
			extensionDevelopmentPath: environmentService.extensionDevelopmentPath,
			extensionTestsPath: environmentService.extensionTestsPath,
			enableProposedApi: false,
		},
		contextService: {
			workspace: contextService.getWorkspace()
		},
		extensions: [],
		configuration: configurationService.values(),
		telemetryInfo: {} as any,
	};
	return createApiFactory(initData, extHostThreadService, extensionService as any, contextService);
}
