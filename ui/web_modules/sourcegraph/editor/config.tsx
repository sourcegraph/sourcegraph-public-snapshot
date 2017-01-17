import * as throttle from "lodash/throttle";
import URI from "vs/base/common/uri";
import { ICodeEditor } from "vs/editor/browser/editorBrowser";
import { EmbeddedCodeEditorWidget } from "vs/editor/browser/widget/embeddedCodeEditorWidget";
import { ICursorSelectionChangedEvent, IRange } from "vs/editor/common/editorCommon";
import { ICodeEditorService } from "vs/editor/common/services/codeEditorService";
import { IEditorService } from "vs/platform/editor/common/editor";
import { IWorkspaceContextService } from "vs/platform/workspace/common/workspace";
import { ExplorerView } from "vs/workbench/parts/files/browser/views/explorerView";
import { IWorkbenchEditorService } from "vs/workbench/services/editor/common/editorService";
import { IViewletService } from "vs/workbench/services/viewlet/browser/viewlet";

import { BlobRouteProps, Router } from "sourcegraph/app/router";
import { urlToBlob } from "sourcegraph/blob/routes";
import { URIUtils } from "sourcegraph/core/uri";
import { getEditorInstance, updateEditorInstance } from "sourcegraph/editor/Editor";
import { WorkbenchEditorService } from "sourcegraph/workbench/overrides/editorService";
import { Services } from "sourcegraph/workbench/services";

/**
 * The currently displayed resource.
 *
 * DEVELOPER NOTE:
 *
 * Workbench state reflects a resource identified by a URI. E.g.
 *   git://github.com/gorilla/mux#mux.go
 *
 * In the example above, the "workspace" is github.com/gorilla/mux,
 * and the "path" is mux.go.
 *
 * The current workbench state is partially managed by VS Code,
 * and partially by the outer React application in which we embed it.
 *
 * For example, when a user does a J2D in the VS Code editor, VS Code
 * will handle updating state. We *also* need to signal to the
 * outer React application that VS Code state has changed. Some outer
 * components may be interested in the fact that the user is now viewing
 * a new file. We signal to the outer react application these facts through
 * react-router, by updating the URL.
 *
 * But what about when the URL changes as the result of some action outside
 * VS Code. For example, clicking some react component. Or hitting "back" in
 * the browser? Usually when these happen, the application root will
 * get signalled by react-route and receive the new URL properties, which it
 * passes down to its children, recursively, letting them (re)draw. We do
 * the same thing in spirit for VS Code, only since it is not a React component,
 * we use a VS Code API to cause redrawing.
 *
 * Caveat:
 *
 * There is some (unexplained) sensitivity in this data flow diagram:
 *
 * React App (listening to URL changes via react-router): -----|
 *                                                             |----> current state
 * VS Code (listening to user actions, like J2D, hover): ------|
 *
 * If the two action sources are updating state at the same time it can cause problems, but *only
 * observed when trying to open the same resource*. To workaround this, we track the the currently
 * displayed resource.
 */
let currResource: URI | null;

export function unmountWorkbench(): void {
	currResource = null;
}

/**
 * syncEditorWithRouterProps forces the editor model to match current URL blob properties.
 */
export function syncEditorWithRouterProps(blobProps: BlobRouteProps): void {
	const {repo, rev, path} = blobProps;
	const resource = URIUtils.pathInRepo(repo, rev, path);
	const editorService = Services.get(IWorkbenchEditorService) as IWorkbenchEditorService;
	if (currResource && currResource.toString() === resource.toString()) {
		return;
	}
	currResource = resource;
	editorService.openEditor({ resource }).then(() => {
		updateFileTree(resource);
		updateEditorAfterURLChange(blobProps.selection);
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
}

/**
 * registerEditorCallbacks attaches custom Sourcegraph handling to the workbench editor lifecycle.
 */
export function registerEditorCallbacks(router: Router): void {
	const codeEditorService = Services.get(ICodeEditorService) as ICodeEditorService;
	codeEditorService.onCodeEditorAdd(updateEditor);

	const editorService = Services.get(IEditorService) as WorkbenchEditorService;
	editorService.onDidOpenEditor(uri => editorOpened(uri, router));
}

/**
 * editorOpened is called whenever the view of the file changes from an action. E.g:
 *  - page load
 *  - file in explorer selected
 *  - jump to definition
 */
function editorOpened(resource: URI, router: Router): void {
	if (currResource && currResource.toString() === resource.toString()) {
		return;
	}
	currResource = resource;
	updateFileTree(resource);
	let {repo, rev, path} = URIUtils.repoParams(resource);
	if (rev === "HEAD") {
		rev = null;
	}
	router.push(urlToBlob(repo, rev, path));
}

async function updateFileTree(resource: URI): Promise<void> {
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
	await view.select(resource);
	const scrollPos = privateView.tree.getRelativeTop(fileStat);
	if (scrollPos > 1 || scrollPos < 0) {
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
	const startLine = e.selection.startLineNumber;
	const endLine = e.selection.endLineNumber;

	let lineHash: string;
	if (startLine !== endLine) {
		lineHash = "#L" + startLine + "-" + endLine;
	} else {
		lineHash = "#L" + startLine;
	}

	// Circumvent react-router to avoid a jarring jump to the anchor position.
	history.replaceState({}, "", window.location.pathname + lineHash);
}
