import Event from "vs/base/common/event";
import { TPromise } from "vs/base/common/winjs.base";
import { DynamicStandaloneServices } from "vs/editor/browser/standalone/standaloneServices";
import { ITextModelResolverService } from "vs/editor/common/services/resolverService";
import { IBackupService } from "vs/platform/backup/common/backup";
import { IEnvironmentService } from "vs/platform/environment/common/environment";
import { IInstantiationService } from "vs/platform/instantiation/common/instantiation";
import { ServiceCollection } from "vs/platform/instantiation/common/serviceCollection";
import { IIntegrityService, IntegrityTestResult } from "vs/platform/integrity/common/integrity";
import { ILifecycleService } from "vs/platform/lifecycle/common/lifecycle";
import { IMessageService } from "vs/platform/message/common/message";
import { IWindowService, IWindowsService } from "vs/platform/windows/common/windows";
import { WorkbenchMessageService } from "vs/workbench/services/message/browser/messageService";
import { TextModelResolverService } from "vs/workbench/services/textmodelResolver/common/textModelResolverService";
import { IThemeService } from "vs/workbench/services/themes/common/themeService";
import { IUntitledEditorService, UntitledEditorService } from "vs/workbench/services/untitled/common/untitledEditorService";

export function setupServices(domElement: HTMLDivElement): ServiceCollection {
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

	return services;
}

class LifecycleService implements ILifecycleService {

	_serviceBrand: any;

	willShutdown: boolean = false;

	onWillShutdown(): any {
		//
	}

	onShutdown(): any {
		//
	}

}

class EnvironmentService implements IEnvironmentService {

	_serviceBrand: any;

	appSettingsHome: string = "home";
}

class WindowService implements IWindowService {

	_serviceBrand: any;

	getCurrentWindowId(): number {
		return 2;
	}

}

class WindowsService implements IWindowsService {

	_serviceBrand: any;

}

class IntegrityService implements IIntegrityService {

	_serviceBrand: any;

	isPure(): TPromise<IntegrityTestResult> {
		return TPromise.wrap({
			isPure: true,
			proof: [1, 2],
		} as any);
	}

}

class BackupService implements IBackupService {

	_serviceBrand: any;

	getBackupPath(): TPromise<string> {
		return TPromise.wrap("some backup path");
	}
}

class ThemeService implements IThemeService {

	_serviceBrand: any;

	onDidColorThemeChange(): Event<string> {
		return "foo" as any;
	}

	getColorTheme(): string {
		return "vs-dark";
	}

}
