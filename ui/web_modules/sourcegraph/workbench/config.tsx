import { IModeService } from "vs/editor/common/services/modeService";
import { ITextModelResolverService } from "vs/editor/common/services/resolverService";
import { ContextMenuController } from "vs/editor/contrib/contextmenu/browser/contextmenu";
import { IInstantiationService } from "vs/platform/instantiation/common/instantiation";
import { ServiceCollection } from "vs/platform/instantiation/common/serviceCollection";
import { Registry } from "vs/platform/platform";
import { IStorageService, StorageScope } from "vs/platform/storage/common/storage";
import { EditorGroupsControl } from "vs/workbench/browser/parts/editor/editorGroupsControl";
import { Extensions as viewKey, ViewletRegistry } from "vs/workbench/browser/viewlet";
import { FileRenderer } from "vs/workbench/parts/files/browser/views/explorerViewer";
import { VIEWLET_ID } from "vs/workbench/parts/files/common/files";
import { StorageService } from "vs/workbench/services/storage/common/storageService";
import { ITextFileService } from "vs/workbench/services/textfile/common/textfiles";

import { layout } from "sourcegraph/components/utils";
import { TextModelContentProvider } from "sourcegraph/workbench/overrides/resolverService";

// Set the height of files in the file tree explorer.
(FileRenderer as any).ITEM_HEIGHT = 30;

// Set the height of the blob title.
(EditorGroupsControl as any).EDITOR_TITLE_HEIGHT = layout.EDITOR_TITLE_HEIGHT;

export function configurePreStartup(services: ServiceCollection): void {
	const instantiationService = services.get(IInstantiationService) as IInstantiationService;

	const storageService = instantiationService.createInstance((StorageService as any), window.localStorage, window.localStorage) as IStorageService;
	services.set(IStorageService, storageService);

	const viewReg = (Registry.as(viewKey.Viewlets) as ViewletRegistry);
	viewReg.setDefaultViewletId(VIEWLET_ID);

	const key = "workbench.sidebar.width";
	storageService.store(key, 300, StorageScope.GLOBAL);
}

// Workbench overwrites a few services, so we add these services after startup.
export function configurePostStartup(services: ServiceCollection): void {
	const resolver = services.get(ITextModelResolverService) as ITextModelResolverService;
	resolver.registerTextModelContentProvider("git", new TextModelContentProvider(
		services.get(IModeService) as IModeService,
		services.get(ITextFileService) as ITextFileService,
	));

	(ContextMenuController.prototype as any)._onContextMenu = () => { /* */ };
}
