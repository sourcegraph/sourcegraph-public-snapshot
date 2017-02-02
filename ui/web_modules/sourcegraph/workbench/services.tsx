import Event from "vs/base/common/event";
import { IDisposable } from "vs/base/common/lifecycle";
import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { StaticServices } from "vs/editor/browser/standalone/standaloneServices";
import { ITextModelResolverService } from "vs/editor/common/services/resolverService";
import { IBackupService } from "vs/platform/backup/common/backup";
import { IConfigurationService } from "vs/platform/configuration/common/configuration";
import { IEnvironmentService } from "vs/platform/environment/common/environment";
import { IExtensionService } from "vs/platform/extensions/common/extensions";
import { ServicesAccessor } from "vs/platform/instantiation/common/instantiation";
import { ServiceCollection } from "vs/platform/instantiation/common/serviceCollection";
import { IIntegrityService, IntegrityTestResult } from "vs/platform/integrity/common/integrity";
import { ILifecycleService } from "vs/platform/lifecycle/common/lifecycle";
import { IMessageService } from "vs/platform/message/common/message";
import "vs/platform/opener/browser/opener.contribution";
import { ISearchService } from "vs/platform/search/common/search";
import { IWindowService, IWindowsService } from "vs/platform/windows/common/windows";
import { IWorkspaceContextService, WorkspaceContextService } from "vs/platform/workspace/common/workspace";
import { ITreeExplorerService } from "vs/workbench/parts/explorers/common/treeExplorerService";
import { IWorkspaceConfigurationService } from "vs/workbench/services/configuration/common/configuration";
import { WorkbenchMessageService } from "vs/workbench/services/message/browser/messageService";
import { ITextFileService } from "vs/workbench/services/textfile/common/textfiles";
import { TextModelResolverService } from "vs/workbench/services/textmodelResolver/common/textModelResolverService";
import { IThemeService } from "vs/workbench/services/themes/common/themeService";
import { IThreadService } from "vs/workbench/services/thread/common/threadService";
import { IUntitledEditorService, UntitledEditorService } from "vs/workbench/services/untitled/common/untitledEditorService";
import { IWindowIPCService } from "vs/workbench/services/window/electron-browser/windowService";

import { MainThreadService } from "sourcegraph/ext/mainThreadService";
import { ConfigurationService, WorkspaceConfigurationService } from "sourcegraph/workbench/ConfigurationService";
import { EnvironmentService } from "sourcegraph/workbench/environmentService";
import { ExtensionService } from "sourcegraph/workbench/extensionService";
import { standaloneServices } from "sourcegraph/workbench/standaloneServices";
import { NoopDisposer } from "sourcegraph/workbench/utils";

export let Services: ServicesAccessor;

// Setup services for the workbench. A lot of the ones required by Workbench
// aren't necessary for Sourcegraph at this point. For instance,
// EnvironmentService isn't something we need because a user will not have a
// home directory on Sourcegraph.

// Others, like ThemeService, will probably be implemented someday, so users
// can customize color themes. When they are implemented, we can either use the
// VSCode ones and override some methods, or we can write our own from scratch.
export function setupServices(domElement: HTMLDivElement, workspace: URI): ServiceCollection {
	const [services, instantiationService] = StaticServices.init({});

	const set = (identifier, impl) => {
		const instance = instantiationService.createInstance(impl);
		services.set(identifier, instance);
	};

	standaloneServices(domElement, services);

	// Override standalone WorkspaceContextService immediately so
	// that services below that depend on it use our overridden
	// service.
	services.set(IWorkspaceContextService, new WorkspaceContextService({
		resource: workspace,
	}));

	set(IUntitledEditorService, UntitledEditorService);
	set(ILifecycleService, LifecycleService);
	set(IEnvironmentService, EnvironmentService);
	set(IWindowService, WindowService);
	set(IWindowsService, DummyService);
	set(IIntegrityService, IntegrityService);
	set(IBackupService, BackupService);
	set(IMessageService, WorkbenchMessageService);
	set(IThemeService, ThemeService);
	set(IWindowIPCService, DummyService);
	set(ITextFileService, DummyService);
	set(ITextModelResolverService, TextModelResolverService);
	set(IConfigurationService, ConfigurationService);
	set(IWorkspaceConfigurationService, WorkspaceConfigurationService);
	set(IThreadService, MainThreadService);
	set(IExtensionService, ExtensionService);

	// These services are depended on by the extension host but are
	// not actually used yet.
	set(ISearchService, function (): void {/* noop */ } as any);
	set(ITreeExplorerService, function (): void { /* noop */ } as any);

	Services = services;

	return services;
}

class DummyService { }

class LifecycleService {

	willShutdown: boolean = false;

	onWillShutdown(): IDisposable {
		return NoopDisposer;
	}

	onShutdown(): any {
		//
	}

}

class WindowService {

	getCurrentWindowId(): number {
		return 1;
	}

}

class IntegrityService {

	isPure(): TPromise<IntegrityTestResult> {
		return TPromise.wrap({
			isPure: true,
		} as any);
	}

}

class BackupService {

	getBackupPath(): TPromise<string> {
		return TPromise.wrap("some backup path");
	}
}

class ThemeService {

	onDidColorThemeChange(): Event<string> {
		return NoopDisposer as any;
	}

	getColorTheme(): string {
		return "vs-dark";
	}

}
