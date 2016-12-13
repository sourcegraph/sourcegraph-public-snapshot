import { IConfigurationService } from "vs/platform/configuration/common/configuration";
import { ServiceCollection } from "vs/platform/instantiation/common/serviceCollection";
import { Registry } from "vs/platform/platform";
import { Extensions as viewKey, ViewletRegistry } from "vs/workbench/browser/viewlet";
import { VIEWLET_ID } from "vs/workbench/parts/files/common/files";

export function configureServices(services: ServiceCollection): void {
	const configsvc = services.get(IConfigurationService) as IConfigurationService;
	configsvc["_config"] = config;
	const viewReg = (Registry.as(viewKey.Viewlets) as ViewletRegistry);
	viewReg.setDefaultViewletId(VIEWLET_ID);
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
