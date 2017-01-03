import { StandaloneCommandService } from "vs/editor/browser/standalone/simpleServices";
import { IMenuService } from "vs/platform/actions/common/actions";
import { MenuService } from "vs/platform/actions/common/menuService";
import { ICommandService } from "vs/platform/commands/common/commands";
import { ContextKeyService } from "vs/platform/contextkey/browser/contextKeyService";
import { IContextKeyService } from "vs/platform/contextkey/common/contextkey";
import { ContextMenuService } from "vs/platform/contextview/browser/contextMenuService";
import { IContextMenuService, IContextViewService } from "vs/platform/contextview/browser/contextView";
import { ContextViewService } from "vs/platform/contextview/browser/contextViewService";
import { IInstantiationService } from "vs/platform/instantiation/common/instantiation";
import { ServiceCollection } from "vs/platform/instantiation/common/serviceCollection";

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
