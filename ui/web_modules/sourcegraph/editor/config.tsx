import * as throttle from "lodash/throttle";
import URI from "vs/base/common/uri";
import { ICodeEditor } from "vs/editor/browser/editorBrowser";
import { EmbeddedCodeEditorWidget } from "vs/editor/browser/widget/embeddedCodeEditorWidget";
import { ICursorSelectionChangedEvent, IRange } from "vs/editor/common/editorCommon";
import { ICodeEditorService } from "vs/editor/common/services/codeEditorService";
import { ITextModelResolverService } from "vs/editor/common/services/resolverService";
import { IWorkspaceContextService } from "vs/platform/workspace/common/workspace";
import { ResourceEditorInput } from "vs/workbench/common/editor/resourceEditorInput";
import { ExplorerView } from "vs/workbench/parts/files/browser/views/explorerView";
import { IWorkbenchEditorService } from "vs/workbench/services/editor/common/editorService";
import { IViewletService } from "vs/workbench/services/viewlet/browser/viewlet";

import { abs, getRoutePattern } from "sourcegraph/app/routePatterns";
import { Router } from "sourcegraph/app/router";
import { __getRouterForWorkbenchOnly } from "sourcegraph/app/router";
import { AbsoluteLocation } from "sourcegraph/core/rangeOrPosition";
import { RangeOrPosition } from "sourcegraph/core/rangeOrPosition";
import { URIUtils } from "sourcegraph/core/uri";
import { getEditorInstance, updateEditorInstance } from "sourcegraph/editor/Editor";
import { GoToDefinitionAction } from "sourcegraph/workbench/info/action";
import { WorkbenchEditorService } from "sourcegraph/workbench/overrides/editorService";
import { Services } from "sourcegraph/workbench/services";

/**
 * syncEditorWithRouterProps forces the editor model to match current URL blob properties.
 */
export function syncEditorWithRouterProps(location: AbsoluteLocation): void {
	const { repo, commitID, path, selection } = location;
	const resource = URIUtils.pathInRepo(repo, commitID, path);
	const editorService = Services.get(IWorkbenchEditorService) as WorkbenchEditorService;
	const resolverService = Services.get(ITextModelResolverService);
	const editorInput = new ResourceEditorInput("", "", resource, resolverService);
	editorService.openEditorWithoutURLChange(editorInput).then(() => {
		updateEditorAfterURLChange(selection);
	});
}

function updateEditorAfterURLChange(sel: IRange): void {
	// TODO restore serialized view state.
	if (!sel) {
		return;
	}

	const editor = getEditorInstance();
	editor.setSelection(sel);
	editor.revealRangeInCenter(sel);
	editor.getAction(GoToDefinitionAction.ID).run();
}

/**
 * registerEditorCallbacks attaches custom Sourcegraph handling to the workbench editor lifecycle.
 */
export function registerEditorCallbacks(router: Router): void {
	const codeEditorService = Services.get(ICodeEditorService) as ICodeEditorService;
	codeEditorService.onCodeEditorAdd(updateEditor);
}

export async function updateFileTree(resource: URI): Promise<void> {
	const viewletService = Services.get(IViewletService) as IViewletService;
	const viewlet = viewletService.getActiveViewlet();
	if (!viewlet) {
		return;
	}

	const view = viewlet["explorerView"];
	if (!(view instanceof ExplorerView)) {
		throw new Error("Type Error: Expected viewlet to have type ExplorerView");
	}

	const workspaceService = Services.get(IWorkspaceContextService) as IWorkspaceContextService;
	const newWorkspace = resource.with({ fragment: "" });
	if (workspaceService.getWorkspace().resource.toString() !== newWorkspace.toString()) {
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
	editor.onDidChangeCursorSelection(throttle(updateURLHash, 100, { leading: true, trailing: true }));
}

function updateURLHash(e: ICursorSelectionChangedEvent): void {
	const router = __getRouterForWorkbenchOnly();
	const isSymbolUrl = getRoutePattern(router.routes) === abs.goSymbol;
	if (isSymbolUrl) {
		return;
	}

	const sel = RangeOrPosition.fromMonacoRange(e.selection);
	const hash = `#L${sel.toString()}`;
	// Circumvent react-router to avoid a jarring jump to the anchor position.
	history.replaceState({}, "", window.location.pathname + hash);
}
