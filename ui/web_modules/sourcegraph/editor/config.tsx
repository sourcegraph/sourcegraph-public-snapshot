import * as throttle from "lodash/throttle";
import * as React from "react";
import * as ReactDOM from "react-dom";
import URI from "vs/base/common/uri";
import { ICodeEditor } from "vs/editor/browser/editorBrowser";
import { EmbeddedCodeEditorWidget } from "vs/editor/browser/widget/embeddedCodeEditorWidget";
import { CursorChangeReason, ICursorSelectionChangedEvent, IRange } from "vs/editor/common/editorCommon";
import { ICodeEditorService } from "vs/editor/common/services/codeEditorService";
import { ITextModelResolverService } from "vs/editor/common/services/resolverService";
import { IFileService } from "vs/platform/files/common/files";
import { IWorkspaceContextService } from "vs/platform/workspace/common/workspace";
import { ResourceEditorInput } from "vs/workbench/common/editor/resourceEditorInput";
import { ExplorerView } from "vs/workbench/parts/files/browser/views/explorerView";
import { IWorkbenchEditorService } from "vs/workbench/services/editor/common/editorService";
import { IViewletService } from "vs/workbench/services/viewlet/browser/viewlet";

import { abs, getRoutePattern } from "sourcegraph/app/routePatterns";
import { Router } from "sourcegraph/app/router";
import { __getRouterForWorkbenchOnly } from "sourcegraph/app/router";
import { urlToBlobRange } from "sourcegraph/blob/routes";
import { FlexContainer } from "sourcegraph/components";
import { colors, typography, whitespace } from "sourcegraph/components/utils";
import { AbsoluteLocation, RangeOrPosition } from "sourcegraph/core/rangeOrPosition";
import { URIUtils } from "sourcegraph/core/uri";
import { getEditorInstance, updateEditorInstance } from "sourcegraph/editor/Editor";
import { GoToDefinitionAction } from "sourcegraph/workbench/info/action";
import { WorkbenchEditorService } from "sourcegraph/workbench/overrides/editorService";
import { Services } from "sourcegraph/workbench/services";
import { prettifyRev } from "sourcegraph/workbench/utils";

/**
 * syncEditorWithRouterProps forces the editor model to match current URL blob properties.
 */
export async function syncEditorWithRouterProps(location: AbsoluteLocation): Promise<void> {
	const { repo, commitID, path } = location;
	const resource = URIUtils.pathInRepo(repo, commitID, path);
	const fileStat = await Services.get(IFileService).resolveFile(resource);
	if (fileStat.isDirectory) {
		renderDirectoryContent();
		updateFileTree(resource);
	} else {
		renderFileEditor(location);
	}
}

/**
 * renderEditor opens the editor for a file.
 */
function renderFileEditor(location: AbsoluteLocation): void {
	const { repo, commitID, path, selection } = location;
	const resource = URIUtils.pathInRepo(repo, commitID, path);
	const editorService = Services.get(IWorkbenchEditorService) as WorkbenchEditorService;
	const resolverService = Services.get(ITextModelResolverService);
	const editorInput = new ResourceEditorInput("", "", resource, resolverService);
	editorService.openEditorWithoutURLChange(editorInput).then(() => {
		updateEditorAfterURLChange(selection);
	});
}

/**
 * renderRootContent displays a welcome message when a user is viewing the root of or a directory in a repo.
 */
function renderDirectoryContent(): void {
	// We don't need or want the editor to be open when displaying the content for a directory.
	const editorService = Services.get(IWorkbenchEditorService) as WorkbenchEditorService;
	editorService.closeAllEditors();

	const el = document.getElementById("workbench.parts.editor");
	if (!el) {
		throw new Error("Expected workbench.parts.editor element to exist.");
	}
	const container = el.firstChild;
	if (!container) {
		throw new Error("Expected workbench.parts.editor to have a child.");
	}

	const node = document.createElement("div");
	node.style.width = "100%";
	node.style.height = "100%";
	container.appendChild(node);

	const style = {
		fontFamily: typography.fontStack.sansSerif,
		color: colors.white(),
		margin: whitespace[2],
		textAlign: "center",
	};

	const keyboardShortcutStyle = {
		backgroundColor: colors.blueGrayD1(),
		borderRadius: "3px",
		padding: "2px 5px",
	};
	const content = <FlexContainer direction="top_bottom" justify="center" items="center" style={{
		width: "100%",
		height: "100%",
		padding: whitespace[2],
		paddingBottom: whitespace[6],
	}}>
		<div id="directory_help_message" style={{ ...style, ...typography.size[4] }}>Start by going to a file or hit <span style={keyboardShortcutStyle}>/</span > to search for a symbol.</div>
	</FlexContainer>;
	ReactDOM.render(content, node);
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
	editor.onDidChangeCursorSelection(throttle(updateURLHash, 100, { leading: true, trailing: true }));
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

		let query = "";
		// Keep query param for zap when selecting lines.
		if (router.location.query["tmpZapRef"]) {
			query = `?tmpZapRef=${router.location.query["tmpZapRef"]}`;
		}

		// Circumvent react-router to avoid a jarring jump to the anchor position.
		history.replaceState({}, "", window.location.pathname + query + hash);
	}
}
