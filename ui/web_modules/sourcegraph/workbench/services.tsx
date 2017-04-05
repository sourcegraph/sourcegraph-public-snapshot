import { IDisposable } from "vs/base/common/lifecycle";
import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { StaticServices } from "vs/editor/browser/standalone/standaloneServices";
import { IBackupService } from "vs/platform/backup/common/backup";
import { IConfigurationService } from "vs/platform/configuration/common/configuration";
import { IEnvironmentService } from "vs/platform/environment/common/environment";
import { IExtensionService } from "vs/platform/extensions/common/extensions";
import { IFileService } from "vs/platform/files/common/files";
import { ServicesAccessor } from "vs/platform/instantiation/common/instantiation";
import { ServiceCollection } from "vs/platform/instantiation/common/serviceCollection";
import { IIntegrityService, IntegrityTestResult } from "vs/platform/integrity/common/integrity";
import { ILifecycleService } from "vs/platform/lifecycle/common/lifecycle";
import { IChoiceService, IMessageService } from "vs/platform/message/common/message";
import "vs/platform/opener/browser/opener.contribution";
import { ISearchService } from "vs/platform/search/common/search";
import { IWindowService, IWindowsService } from "vs/platform/windows/common/windows";
import { IWorkspace, IWorkspaceContextService, IWorkspaceRevState, WorkspaceContextService } from "vs/platform/workspace/common/workspace";
import { EditorPart } from "vs/workbench/browser/parts/editor/editorPart";
import { ITreeExplorerService } from "vs/workbench/parts/explorers/common/treeExplorerService";
import { IPreferencesService } from "vs/workbench/parts/preferences/common/preferences";
import { IBackupFileService } from "vs/workbench/services/backup/common/backup";
import { IWorkspaceConfigurationService } from "vs/workbench/services/configuration/common/configuration";
import { IEditorGroupService } from "vs/workbench/services/group/common/groupService";
import { MessageService } from "vs/workbench/services/message/electron-browser/messageService";
import { IPartService } from "vs/workbench/services/part/common/partService";
import { IThreadService } from "vs/workbench/services/thread/common/threadService";
import { IUntitledEditorService, UntitledEditorService } from "vs/workbench/services/untitled/common/untitledEditorService";
import { IWindowIPCService } from "vs/workbench/services/window/electron-browser/windowService";

import { MainThreadService } from "sourcegraph/ext/mainThreadService";
import { ConfigurationService, WorkspaceConfigurationService } from "sourcegraph/workbench/ConfigurationService";
import { EnvironmentService } from "sourcegraph/workbench/environmentService";
import { ExtensionService } from "sourcegraph/workbench/extensionService";
import { FileService } from "sourcegraph/workbench/overrides/fileService";
import { SearchService } from "sourcegraph/workbench/searchService";
import { standaloneServices } from "sourcegraph/workbench/standaloneServices";
import { setContextService } from "sourcegraph/workbench/utils";

export const NoopDisposer = { dispose: () => {/* */ } };

export let Services: ServicesAccessor;

// Setup services for the workbench. A lot of the ones required by Workbench
// aren't necessary for Sourcegraph at this point. For instance,
// EnvironmentService isn't something we need because a user will not have a
// home directory on Sourcegraph.

// Others, like ThemeService, will probably be implemented someday, so users
// can customize color themes. When they are implemented, we can either use the
// VSCode ones and override some methods, or we can write our own from scratch.
export function setupServices(domElement: HTMLDivElement, workspace: URI, revState?: IWorkspaceRevState): ServiceCollection {
	const [services, instantiationService] = StaticServices.init({});

	const set = (identifier, impl) => {
		const instance = instantiationService.createInstance(impl);
		services.set(identifier, instance);
	};

	set(IExtensionService, ExtensionService);

	standaloneServices(domElement, services);

	// Override standalone WorkspaceContextService immediately so
	// that services below that depend on it use our overridden
	// service.
	services.set(IWorkspaceContextService, new WorkspaceContextService({
		resource: workspace,
		revState,
	}));

	set(IUntitledEditorService, UntitledEditorService);
	set(ILifecycleService, LifecycleService);
	set(IEnvironmentService, EnvironmentService);
	set(IWindowService, WindowService);
	set(IWindowsService, DummyService);
	set(IIntegrityService, IntegrityService);
	set(IBackupService, BackupService);
	set(IBackupFileService, function (): void { /* noop */ } as any);

	set(IWindowIPCService, DummyService);
	set(IPartService, DummyService);

	const messageService = instantiationService.createInstance(MessageService, domElement);
	services.set(IMessageService, messageService);
	services.set(IChoiceService, messageService);

	const editorPart = instantiationService.createInstance(EditorPart, "workbench.parts.editor", false);
	services.set(IEditorGroupService, editorPart);
	set(IConfigurationService, ConfigurationService);
	set(IWorkspaceConfigurationService, WorkspaceConfigurationService);
	set(IThreadService, MainThreadService);
	set(ISearchService, SearchService);
	// These services are depended on by the extension host but are
	// not actually used yet.
	set(ITreeExplorerService, DummyService);
	set(IPreferencesService, DummyService);
	set(IFileService, FileService);

	Services = services;

	setContextService(Services.get(IWorkspaceContextService));

	return services;
}

export function setWorkspace(workspace: IWorkspace): void {
	if (workspace.revState && workspace.revState.zapRef && !/^branch\//.test(workspace.revState.zapRef)) {
		throw new Error(`invalid Zap ref: ${JSON.stringify(workspace.revState.zapRef)} (no 'branch/' prefix)`);
	}

	const contextService = Services.get(IWorkspaceContextService);
	return contextService.setWorkspace(workspace);
}

export function registerWorkspace(workspace: IWorkspace): void {
	if (workspace.revState && workspace.revState.zapRef && !/^branch\//.test(workspace.revState.zapRef)) {
		throw new Error(`invalid Zap ref: ${JSON.stringify(workspace.revState.zapRef)} (no 'branch/' prefix)`);
	}

	const contextService = Services.get(IWorkspaceContextService);
	return contextService.registerWorkspace(workspace);
}

export function onWorkspaceUpdated(listener: ((workspace: IWorkspace) => any)): IDisposable {
	const contextService = Services.get(IWorkspaceContextService);
	return contextService.onWorkspaceUpdated(listener);
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

	getRecentlyOpen(): TPromise<{ files: string[], folders: string[] }> {
		return TPromise.as({ files: [], folders: [] });
	}

	isMaximized(): TPromise<boolean> {
		return TPromise.as(false);
	}

	maximizeWindow(): TPromise<void> {
		return TPromise.as(void 0);
	}

	unmaximizeWindow(): TPromise<void> {
		return TPromise.as(void 0);
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
