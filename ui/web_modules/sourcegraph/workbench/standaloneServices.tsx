import { CodeEditorServiceImpl } from "vs/editor/browser/services/codeEditorServiceImpl";
import { StandaloneCommandService } from "vs/editor/browser/standalone/simpleServices";
import { ICodeEditorService } from "vs/editor/common/services/codeEditorService";
import { IEditorWorkerService } from "vs/editor/common/services/editorWorkerService";
import { EditorWorkerServiceImpl } from "vs/editor/common/services/editorWorkerServiceImpl";
import { IModelService } from "vs/editor/common/services/modelService";
import { ModelServiceImpl } from "vs/editor/common/services/modelServiceImpl";
import { IModeService } from "vs/editor/common/services/modeService";
import { MainThreadModeServiceImpl } from "vs/editor/common/services/modeServiceImpl";
import { IMenuService } from "vs/platform/actions/common/actions";
import { MenuService } from "vs/platform/actions/common/menuService";
import { ICommandService } from "vs/platform/commands/common/commands";
import { IConfigurationService } from "vs/platform/configuration/common/configuration";
import { ContextKeyService } from "vs/platform/contextkey/browser/contextKeyService";
import { IContextKeyService } from "vs/platform/contextkey/common/contextkey";
import { ContextMenuService } from "vs/platform/contextview/browser/contextMenuService";
import { IContextMenuService, IContextViewService } from "vs/platform/contextview/browser/contextView";
import { ContextViewService } from "vs/platform/contextview/browser/contextViewService";
import { IEventService } from "vs/platform/event/common/event";
import { EventService } from "vs/platform/event/common/eventService";
import { IExtensionService } from "vs/platform/extensions/common/extensions";
import { IInstantiationService, ServiceIdentifier } from "vs/platform/instantiation/common/instantiation";
import { InstantiationService } from "vs/platform/instantiation/common/instantiationService";
import { ServiceCollection } from "vs/platform/instantiation/common/serviceCollection";
import { IKeybindingService } from "vs/platform/keybinding/common/keybinding";
import { IMarkerService } from "vs/platform/markers/common/markers";
import { MarkerService } from "vs/platform/markers/common/markerService";
import { IMessageService } from "vs/platform/message/common/message";
import { IProgressService } from "vs/platform/progress/common/progress";
import { IStorageService, NullStorageService } from "vs/platform/storage/common/storage";
import { ITelemetryService, NullTelemetryService } from "vs/platform/telemetry/common/telemetry";
import { IWorkspaceContextService, WorkspaceContextService } from "vs/platform/workspace/common/workspace";

export function standaloneServices(container: HTMLElement, services: ServiceCollection): void {
	const instantiationService = services.get(IInstantiationService) as IInstantiationService;

	const set = (identifier, impl, arg?) => {
		const instance = instantiationService.createInstance(impl, arg);
		services.set(identifier, instance);
	};

	set(IContextKeyService, ContextKeyService);
	set(ICommandService, StandaloneCommandService);
	set(IContextViewService, ContextViewService, container);
	set(IContextMenuService, ContextMenuService);
	set(IMenuService, MenuService);
}
