import * as throttle from "lodash/throttle";

import { IDisposable } from "vs/base/common/lifecycle";
import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { ICodeEditor } from "vs/editor/browser/editorBrowser";
import { EmbeddedCodeEditorWidget } from "vs/editor/browser/widget/embeddedCodeEditorWidget";
import { CursorChangeReason, ICursorSelectionChangedEvent, IRange } from "vs/editor/common/editorCommon";
import { DefinitionProviderRegistry, HoverProviderRegistry, ReferenceProviderRegistry } from "vs/editor/common/modes";
import { ICodeEditorService } from "vs/editor/common/services/codeEditorService";
import { ITextModelResolverService } from "vs/editor/common/services/resolverService";
import { IFileService } from "vs/platform/files/common/files";
import { IQuickOpenService } from "vs/platform/quickOpen/common/quickOpen";
import { IWorkspaceContextService } from "vs/platform/workspace/common/workspace";
import { DiffEditorInput } from "vs/workbench/common/editor/diffEditorInput";
import { ResourceEditorInput } from "vs/workbench/common/editor/resourceEditorInput";
import { ExplorerView } from "vs/workbench/parts/files/browser/views/explorerView";
import { IWorkbenchEditorService } from "vs/workbench/services/editor/common/editorService";
import { IViewletService } from "vs/workbench/services/viewlet/browser/viewlet";
import { IPartService, Parts } from "vscode/src/vs/workbench/services/part/common/partService";

import { abs, getRoutePattern } from "sourcegraph/app/routePatterns";
import { __getRouterForWorkbenchOnly } from "sourcegraph/app/router";
import { urlToBlobRange } from "sourcegraph/blob/routes";
import { AbsoluteLocation, RangeOrPosition } from "sourcegraph/core/rangeOrPosition";
import { URIUtils } from "sourcegraph/core/uri";
import { renderDirectoryContent, renderNotFoundError } from "sourcegraph/workbench/DirectoryContent";
import { SidebarContribID, SidebarContribution } from "sourcegraph/workbench/info/contrib";
import { getEditorInstance, updateEditorInstance } from "sourcegraph/workbench/overrides/editorService";
import { WorkbenchEditorService } from "sourcegraph/workbench/overrides/editorService";
import { Services, registerWorkspace, setWorkspace } from "sourcegraph/workbench/services";
import { getCurrentWorkspace, getGitBaseResource, getURIContext, getWorkspaceForResource, prettifyRev } from "sourcegraph/workbench/utils";

/**
 * syncEditorWithRouterProps forces the editor model to match current URL blob properties.
 */
export function syncEditorWithRouterProps(location: AbsoluteLocation): void {
	updateWorkspace(location);
	updateEditorArea(location);
}

export function updateWorkspace(location: AbsoluteLocation): void {
	const { repo, path } = location;
	registerWorkspace({ resource: URIUtils.createResourceURI(repo), revState: location });
	const resource = URIUtils.createResourceURI(repo, path === "" ? undefined : path);
	const currWorkspace = getCurrentWorkspace();
	if (getWorkspaceForResource(resource).toString() !== currWorkspace.resource.toString() || (currWorkspace.revState && currWorkspace.revState.zapRef !== location.zapRef)) {
		setWorkspace({ resource: getWorkspaceForResource(resource), revState: { zapRev: location.zapRev, zapRef: location.zapRef, commitID: location.commitID, branch: location.branch } });
	}
}

export async function updateEditorArea(location: AbsoluteLocation): Promise<void> {
	const { repo, path, selection } = location;
	const resource = URIUtils.createResourceURI(repo, path);

	const fileStat = await Services.get(IFileService).resolveFile(resource);
	if (fileStat.isDirectory) {
		renderDirectoryContent();
		updateFileTree(resource);
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
		renderDiffEditor(getGitBaseResource(resource), resource, selection);
	} else {
		renderFileEditor(resource, selection);
	}
}

/**
 * renderEditor opens the editor for a file.
 */
function renderFileEditor(resource: URI, selection: IRange | null): void {
	const editorService = Services.get(IWorkbenchEditorService) as WorkbenchEditorService;
	editorService.openEditorWithoutURLChange(resource, null, { readOnly: false, preserveFocus: true }).then(() => {
		updateEditorAfterURLChange(selection);
	});
}

/**
 * renderEditor opens a diff editor for two files.
 */
function renderDiffEditor(left: URI, right: URI, selection: IRange | null): void {
	const editorService = Services.get(IWorkbenchEditorService) as WorkbenchEditorService;
	const resolverService = Services.get(ITextModelResolverService);
	TPromise.join([editorService.createInput({ resource: left }), editorService.createInput({ resource: right })]).then(inputs => {
		const leftInput = new ResourceEditorInput("", "", left, resolverService);
		const rightInput = new ResourceEditorInput("", "", right, resolverService);
		const diff = new DiffEditorInput("", "", leftInput, rightInput);
		editorService.openEditorWithoutURLChange(right, diff, {}).then(() => {
			updateEditorAfterURLChange(selection);
		});
	});
}

/**
 * isOnZapRev returns whether the user is currently viewing a Zap (not Git) revision.
 */
export function isOnZapRev(): boolean {
	const contextService = Services.get(IWorkspaceContextService) as IWorkspaceContextService;
	return Boolean(contextService.getWorkspace().revState && contextService.getWorkspace().revState!.zapRef);
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
	return disposables;
}

/**
 * toggleQuickopen toggles the quickopen modal state.
 */
export function toggleQuickopen(forceHide?: boolean): void {
	const quickopen = Services.get(IQuickOpenService);
	if (quickOpenShown || forceHide) {
		quickopen.close();
	} else {
		quickopen.show();
	}
}

export async function updateFileTree(resource: URI): Promise<void> {
	const partService = Services.get(IPartService) as IPartService;
	const visible = partService.isVisible(Parts.SIDEBAR_PART);
	if (!visible) {
		return;
	}

	const viewletService = Services.get(IViewletService) as IViewletService;
	let viewlet = viewletService.getActiveViewlet();
	if (!viewlet) {
		viewlet = await new Promise(resolve => {
			viewletService.onDidViewletOpen(resolve);
		}) as any;
	}

	const view = viewlet["explorerView"];
	if (view instanceof ExplorerView) {
		await view.refresh(true);

		const privateView = view as any;
		let root = privateView.root;
		if (!root) {
			await view.refresh();
			root = privateView.root;
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
			await view.refresh(true);

			const privateView = view as any;
			let root = privateView.root;
			if (!root) {
				await view.refresh();
				root = privateView.root;
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
		const { repo, rev, path } = getURIContext(uri);
		const prettyRev = prettifyRev(rev);
		router.push(urlToBlobRange(repo, prettyRev, path, sel.toZeroIndexedRange()));
	} else {
		const hash = `#L${sel.toString()}`;

		// Circumvent react-router to avoid a jarring jump to the anchor position.
		history.replaceState({}, "", window.location.pathname + hash);
	}
}
