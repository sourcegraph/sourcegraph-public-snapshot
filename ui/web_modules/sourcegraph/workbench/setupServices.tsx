import Event from "vs/base/common/event";
import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { DynamicStandaloneServices } from "vs/editor/browser/standalone/standaloneServices";
import { ITextModelResolverService } from "vs/editor/common/services/resolverService";
import { IBackupService } from "vs/platform/backup/common/backup";
import { IConfigurationService } from "vs/platform/configuration/common/configuration";
import { IEnvironmentService } from "vs/platform/environment/common/environment";
import { IInstantiationService } from "vs/platform/instantiation/common/instantiation";
import { ServiceCollection } from "vs/platform/instantiation/common/serviceCollection";
import { IIntegrityService, IntegrityTestResult } from "vs/platform/integrity/common/integrity";
import { ILifecycleService } from "vs/platform/lifecycle/common/lifecycle";
import { IMessageService } from "vs/platform/message/common/message";
import { Registry } from "vs/platform/platform";
import { IWindowService, IWindowsService } from "vs/platform/windows/common/windows";
import { IWorkspaceContextService, WorkspaceContextService } from "vs/platform/workspace/common/workspace";
import { Extensions as viewKey, ViewletRegistry } from "vs/workbench/browser/viewlet";
import { VIEWLET_ID } from "vs/workbench/parts/files/common/files";
import { WorkbenchMessageService } from "vs/workbench/services/message/browser/messageService";
import { TextModelResolverService } from "vs/workbench/services/textmodelResolver/common/textModelResolverService";
import { IThemeService } from "vs/workbench/services/themes/common/themeService";
import { IUntitledEditorService, UntitledEditorService } from "vs/workbench/services/untitled/common/untitledEditorService";


import { ConfigurationService } from "sourcegraph/workbench/config";

// Setup services for the workbench. A lot of the ones required by Workbench
// aren't necessary for Sourcegraph at this point. For instance,
// EnvironmentService isn't something we need because a user will not have a
// home directory on Sourcegraph.

// Others, like ThemeService, will probably be implemented someday, so users
// can customize color themes. When they are implemented, we can either use the
// VSCode ones and override some methods, or we can write our own from scratch.

export function setupServices(domElement: HTMLDivElement, resource: URI): ServiceCollection {
	const dynServices = new DynamicStandaloneServices(domElement, {});
	const services = (dynServices as any)._serviceCollection;
	const instantiationService = services.get(IInstantiationService) as IInstantiationService;

	const set = (identifier, impl) => {
		const instance = instantiationService.createInstance(impl);
		services.set(identifier, instance);
	};

	set(IUntitledEditorService, UntitledEditorService);
	set(ILifecycleService, LifecycleService);
	set(IEnvironmentService, EnvironmentService);
	set(IWindowService, WindowService);
	set(IWindowsService, WindowsService);
	set(IIntegrityService, IntegrityService);
	set(IBackupService, BackupService);
	set(IMessageService, WorkbenchMessageService);
	set(IThemeService, ThemeService);
	set(ITextModelResolverService, TextModelResolverService);
	set(IConfigurationService, ConfigurationService);

	services.set(IWorkspaceContextService, new WorkspaceContextService({
		resource,
	}));

	const viewReg = (Registry.as(viewKey.Viewlets) as ViewletRegistry);
	viewReg.setDefaultViewletId(VIEWLET_ID);

	return services;
}

class LifecycleService {

	_serviceBrand: any;

	willShutdown: boolean = false;

	onWillShutdown(): any {
		//
	}

	onShutdown(): any {
		//
	}

}

class EnvironmentService {

	_serviceBrand: any;

	appSettingsHome: string = "app-settings-home";
}

class WindowService {

	_serviceBrand: any;

	getCurrentWindowId(): number {
		return 1;
	}

}

class WindowsService {

	_serviceBrand: any;

}

class IntegrityService {

	_serviceBrand: any;

	isPure(): TPromise<IntegrityTestResult> {
		return TPromise.wrap({
			isPure: true,
		} as any);
	}

}

class BackupService {

	_serviceBrand: any;

	getBackupPath(): TPromise<string> {
		return TPromise.wrap("some backup path");
	}
}

class ThemeService {

	_serviceBrand: any;

	onDidColorThemeChange(): Event<string> {
		return {} as any;
	}

	getColorTheme(): string {
		return "vs-dark";
	}

}
