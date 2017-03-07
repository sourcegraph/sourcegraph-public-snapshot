import * as throttle from "lodash/throttle";

import { IDisposable } from "vs/base/common/lifecycle";
import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { ICodeEditor } from "vs/editor/browser/editorBrowser";
import { EmbeddedCodeEditorWidget } from "vs/editor/browser/widget/embeddedCodeEditorWidget";
import { CursorChangeReason, ICursorSelectionChangedEvent, IRange } from "vs/editor/common/editorCommon";
import { DefinitionProviderRegistry, HoverProviderRegistry, ReferenceProviderRegistry } from "vs/editor/common/modes";
import { ICodeEditorService } from "vs/editor/common/services/codeEditorService";
import { getCodeEditor } from "vs/editor/common/services/codeEditorService";
import { ITextModelResolverService } from "vs/editor/common/services/resolverService";
import { CommandsRegistry } from "vs/platform/commands/common/commands";
import { IFileService } from "vs/platform/files/common/files";
import { ServicesAccessor } from "vs/platform/instantiation/common/instantiation";
import { IWorkspaceContextService } from "vs/platform/workspace/common/workspace";
import { DiffEditorInput } from "vs/workbench/common/editor/diffEditorInput";
import { ResourceEditorInput } from "vs/workbench/common/editor/resourceEditorInput";
import { ExplorerView } from "vs/workbench/parts/files/browser/views/explorerView";
import { IWorkbenchEditorService } from "vs/workbench/services/editor/common/editorService";
import { IQuickOpenService } from "vs/workbench/services/quickopen/common/quickOpenService";
import { IViewletService } from "vs/workbench/services/viewlet/browser/viewlet";

import { abs, getRoutePattern } from "sourcegraph/app/routePatterns";
import { __getRouterForWorkbenchOnly } from "sourcegraph/app/router";
import { urlToBlobRange } from "sourcegraph/blob/routes";
import { AbsoluteLocation, RangeOrPosition } from "sourcegraph/core/rangeOrPosition";
import { URIUtils } from "sourcegraph/core/uri";
import { getEditorInstance, updateEditorInstance } from "sourcegraph/editor/Editor";
import { renderDirectoryContent, renderNotFoundError } from "sourcegraph/workbench/DirectoryContent";
import { SidebarContribID, SidebarContribution } from "sourcegraph/workbench/info/contrib";
import { onWorkbenchShown } from "sourcegraph/workbench/main";
import { WorkbenchEditorService } from "sourcegraph/workbench/overrides/editorService";
import { Services } from "sourcegraph/workbench/services";
import { prettifyRev } from "sourcegraph/workbench/utils";

/**
 * syncEditorWithRouterProps forces the editor model to match current URL blob properties.
 */
export async function syncEditorWithRouterProps(location: AbsoluteLocation): Promise<void> {
	const { repo, commitID, path, selection } = location;
	const resource = URIUtils.pathInRepo(repo, commitID, path);
	updateFileTree(resource);

	const fileStat = await Services.get(IFileService).resolveFile(resource);
	if (fileStat.isDirectory) {
		renderDirectoryContent();
		return;
	}

	const exists = await Services.get(IFileService).existsFile(resource);
	if (!exists) {
		if (location.zapRef) {
			// Don't render 404 in a zap session yet since the file may have been
			// created by an op.
			return;
		}
		renderNotFoundError();
		return;
	}
	if (location.zapRef) {
		renderDiffEditor(resource.with({ query: `${resource.query}~0` }), resource);
	} else {
		renderFileEditor(resource, selection);
	}
}

function resourceForCurrentEditor(): URI | null {
	const editorService = Services.get(IWorkbenchEditorService) as WorkbenchEditorService;
	const input = editorService.getActiveEditor();
	const editor = getCodeEditor(input);
	if (!editor) { return null; }
	return editor.getModel().uri;
}

/**
 * renderEditor opens the editor for a file.
 */
export function renderFileEditor(resource: URI, selection: IRange | null): void {
	const editorService = Services.get(IWorkbenchEditorService) as WorkbenchEditorService;
	editorService.openEditorWithoutURLChange(resource, null, { readOnly: false }).then(() => {
		updateEditorAfterURLChange(selection);
	});
}

/**
 * renderEditor opens a diff editor for two files.
 */
export function renderDiffEditor(left: URI, right: URI): void {
	const editorService = Services.get(IWorkbenchEditorService) as WorkbenchEditorService;
	const resolverService = Services.get(ITextModelResolverService);
	TPromise.join([editorService.createInput({ resource: left }), editorService.createInput({ resource: right })]).then(inputs => {
		const leftInput = new ResourceEditorInput("", "", left, resolverService);
		const rightInput = new ResourceEditorInput("", "", right, resolverService);
		const diff = new DiffEditorInput("", "", leftInput, rightInput);
		editorService.openEditorWithoutURLChange(right, diff, {});
	});
}

/**
 * toggleEditorDiffMode toggles the mode for the current editor (diff on/off).
 */
export function toggleEditorDiffMode(): void {
	const editorService = Services.get(IWorkbenchEditorService) as WorkbenchEditorService;
	if (editorService.diffMode) {
		changeEditorMode(false);
	} else {
		changeEditorMode(true);
	}
}

/**
 * isOnZapRef returns whether the user is currently viewing a zap ref
 */
export function isOnZapRef(): boolean {
	const contextService = Services.get(IWorkspaceContextService) as IWorkspaceContextService;
	return Boolean(contextService.getWorkspace().revState && contextService.getWorkspace().revState!.zapRef);
}

/**
 * changeEditorMode changes the diff mode for the current editor.
 */
export function changeEditorMode(diff?: boolean): void {
	const uri = resourceForCurrentEditor();
	if (!uri) {
		return;
	}

	const contextService = Services.get(IWorkspaceContextService) as IWorkspaceContextService;
	const revState = contextService.getWorkspace().revState;

	if (diff) {
		const left = uri.with({ query: revState && revState.zapRef ? `${uri.query}~0` : `${uri.query}^` });
		renderDiffEditor(left, uri);
	} else {
		renderFileEditor(uri, null);
	}
	return;
}

function updateEditorAfterURLChange(sel: IRange | null): void {
	// TODO restore scroll position.
	if (!sel) {
		return;
	}

	const editor = getEditorInstance();
	if (!editor) {
		return;
	}
	editor.setSelection(sel);
	editor.revealRangeInCenter(sel);

	// Opening sidebar is a noop until a definition provider is registered.
	// This sidebar ALSO needs hover/reference providers registered to fetch data.
	// The extension host will register providers asynchronously, so wait
	// for registration events before opening the sidebar.
	const providerRegistered = (registry) => {
		return new Promise<void>((resolve, reject) => {
			if (registry.all(editor.getModel()).length === 0) {
				const disposable = registry.onDidChange(() => {
					// assume the change is a registration as needed by the sidebar
					disposable.dispose();
					resolve();
				});
			} else {
				resolve();
			}
		});
	};
	Promise.all([providerRegistered(DefinitionProviderRegistry), providerRegistered(HoverProviderRegistry), providerRegistered(ReferenceProviderRegistry)])
		.then(() => {
			const sidebar = editor.getContribution(SidebarContribID) as SidebarContribution;
			sidebar.openInSidebar();
		});
}

let quickOpenShown = false;

/**
 * registerEditorCallbacks attaches custom Sourcegraph handling to the workbench editor lifecycle.
 */
export function registerEditorCallbacks(): IDisposable[] {
	const disposables: IDisposable[] = [];
	disposables.push(...registerQuickopenListeners(() => quickOpenShown = true, () => quickOpenShown = false));
	const codeEditorService = Services.get(ICodeEditorService) as ICodeEditorService;
	disposables.push(codeEditorService.onCodeEditorAdd(updateEditor));
	return disposables;
}

/**
 * registerQuickopenListeners attaches callbacks which are invoked when a quickopen
 * is shown/closed.
 */
export function registerQuickopenListeners(onShow: () => any, onHide: () => any): IDisposable[] {
	const disposables: IDisposable[] = [];
	const quickOpenService = Services.get(IQuickOpenService) as IQuickOpenService;
	disposables.push(quickOpenService.onShow(onShow));
	disposables.push(quickOpenService.onHide(onHide));
	disposables.push(onWorkbenchShown(shown => !shown && onHide())); // unmounting workbench auto-dismisses quickopen
	return disposables;
}

/**
 * toggleQuickopen toggles the quickopen modal state.
 */
export function toggleQuickopen(): void {
	const quickopen = Services.get(IQuickOpenService);
	if (quickOpenShown) {
		quickopen.close();
	} else {
		quickopen.show();
	}
}

export async function updateFileTree(resource: URI): Promise<void> {
	const viewletService = Services.get(IViewletService) as IViewletService;
	let viewlet = viewletService.getActiveViewlet();
	if (!viewlet) {
		viewlet = await new Promise(resolve => {
			viewletService.onDidViewletOpen(resolve);
		}) as any;
	}

	const view = viewlet["explorerView"];
	if (!(view instanceof ExplorerView)) {
		throw new Error("Type Error: Expected viewlet to have type ExplorerView");
	}

	const workspaceService = Services.get(IWorkspaceContextService) as IWorkspaceContextService;
	const newWorkspace = resource.with({ fragment: "" });
	if (workspaceService.getWorkspace().resource.toString() !== newWorkspace.toString()) {
		// TODO(john): this code doesn't work properly with Zap branch switching
		workspaceService.setWorkspace({ resource: newWorkspace });
		await view.refresh(true);
	}

	const privateView = view as any;
	let root = privateView.getInput();
	if (!root) {
		await view.refresh();
		root = privateView.getInput();
	}
	const fileStat = root.find(resource);
	const treeModel = privateView.tree.model;
	const chain = await treeModel.resolveUnknownParentChain(fileStat);
	chain.forEach((item) => {
		treeModel.expand(item);
	});
	treeModel.expand(fileStat);

	const oldSelection = privateView.tree.getSelection();
	await view.select(resource);
	const scrollPos = privateView.tree.getRelativeTop(fileStat);
	if (scrollPos > 1 || scrollPos < 0 || oldSelection.length === 0) {
		// Item is scrolled off screen
		await view.select(resource, true);
	}
}

function updateEditor(editor: ICodeEditor): void {
	if (editor instanceof EmbeddedCodeEditorWidget) {
		// Don't update the editor instance or the URL hash from the rift view.
		return;
	}
	updateEditorInstance(editor);

	// Listeners
	editor.onDidChangeCursorSelection(throttle(updateURLHash, 200, { leading: true, trailing: true }));
}

function updateURLHash(e: ICursorSelectionChangedEvent): void {
	const router = __getRouterForWorkbenchOnly();
	const isSymbolUrl = getRoutePattern(router.routes) === abs.goSymbol;
	if (isSymbolUrl && e.reason === CursorChangeReason.NotSet) {
		// When landing at a symbol URL, don't update URL.
		return;
	}

	const sel = RangeOrPosition.fromMonacoRange(e.selection);

	if (isSymbolUrl) {
		// When updating selection from a symbol URL, update router location
		// to blob URL.
		const editor = getEditorInstance();
		if (!editor) {
			return;
		}
		const uri = editor.getModel().uri;
		const prettyRev = prettifyRev(uri.query);
		router.push(urlToBlobRange(`${uri.authority}/${uri.path}`, prettyRev || "", uri.fragment, sel.toZeroIndexedRange()));
	} else {
		const hash = `#L${sel.toString()}`;

		// Circumvent react-router to avoid a jarring jump to the anchor position.
		history.replaceState({}, "", window.location.pathname + hash);
	}
}

CommandsRegistry.registerCommand("zap.reference.change", (accessor: ServicesAccessor, ref: string) => {
	// TODO(john): remove
});

CommandsRegistry.registerCommand("zap.status.change", (accessor: ServicesAccessor, isRunning: boolean) => {
	// TODO(john): remove
});
