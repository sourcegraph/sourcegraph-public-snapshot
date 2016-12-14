import URI from "vs/base/common/uri";
import { IModelService } from "vs/editor/common/services/modelService";
import { IModeService } from "vs/editor/common/services/modeService";
import { ITextModelResolverService } from "vs/editor/common/services/resolverService";
import { IConfigurationService } from "vs/platform/configuration/common/configuration";
import { ServiceCollection } from "vs/platform/instantiation/common/serviceCollection";
import { Registry } from "vs/platform/platform";
import { IWorkspaceContextService, WorkspaceContextService } from "vs/platform/workspace/common/workspace";
import { Extensions as viewKey, ViewletRegistry } from "vs/workbench/browser/viewlet";
import { VIEWLET_ID } from "vs/workbench/parts/files/common/files";

import { TextModelContentProvider } from "sourcegraph/editor/resolverService";

export function configureServices(services: ServiceCollection, resource: URI): void {
	const configsvc = services.get(IConfigurationService) as IConfigurationService;
	configsvc["_config"] = config;
	const viewReg = (Registry.as(viewKey.Viewlets) as ViewletRegistry);
	viewReg.setDefaultViewletId(VIEWLET_ID);

	services.set(IWorkspaceContextService, new WorkspaceContextService({
		resource,
	}));
}

export function configurePostStartup(services: ServiceCollection): void {
	const resolver = services.get(ITextModelResolverService) as ITextModelResolverService;
	resolver.registerTextModelContentProvider("git", new TextModelContentProvider(
		services.get(IModelService) as IModelService,
		services.get(IModeService) as IModeService,
	));
}

const config = {
	workbench: {
		quickOpen: {
			closeOnFocusLost: false,
		},
		editor: {
			enablePreview: false,
		},
	},
	explorer: {
		openEditors: {
			visible: 0,
		},
	},
	editor: {},
};
