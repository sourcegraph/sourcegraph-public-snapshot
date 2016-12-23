import * as throttle from "lodash/throttle";
import { ICodeEditor, IContentWidget, IEditorMouseEvent } from "vs/editor/browser/editorBrowser";
import { IModelChangedEvent } from "vs/editor/common/editorCommon";
import { ICursorSelectionChangedEvent } from "vs/editor/common/editorCommon";
import { ICodeEditorService } from "vs/editor/common/services/codeEditorService";
import { IEditorInput } from "vs/platform/editor/common/editor";
import { IWorkspaceContextService } from "vs/platform/workspace/common/workspace";
import { EditorPart } from "vs/workbench/browser/parts/editor/editorPart";
import { IWorkbenchEditorService } from "vs/workbench/services/editor/common/editorService";

import { getBlobPropsFromRouter, getSelectionFromRouter, router } from "sourcegraph/app/router";
import { urlToBlob } from "sourcegraph/blob/routes";
import { URIUtils } from "sourcegraph/core/uri";
import { getEditorInstance, updateEditorInstance } from "sourcegraph/editor/Editor";
import { Services } from "sourcegraph/workbench/services";
import { getResource } from "sourcegraph/workbench/utils";

// forceSyncInProgress tells us if openEditor was called from syncEditorWithRouter or if
// it was called internally by VSCode. In the internal case, we want to update
// the external react router state to the editor's view of the world. In the
// external case, the react router state is already correct, and we want to
// update the editor state to the react router state.
let forceSyncInProgress;

// syncEditorWithURL forces the editor model to match current URL blob properties.
// It only needs to be called in an 'onpopstate' handler, for browser forward & back.
export function syncEditorWithRouter(): void {
	const {repo, rev, path} = getBlobPropsFromRouter();
	const resource = URIUtils.pathInRepo(repo, rev, path);
	const editorService = Services.get(IWorkbenchEditorService) as IWorkbenchEditorService;
	const workspaceService = Services.get(IWorkspaceContextService) as IWorkspaceContextService;
	workspaceService.setWorkspace({ resource: resource.with({ fragment: "" }) });
	forceSyncInProgress = true;
	editorService.openEditor({ resource }).then(() => {
		forceSyncInProgress = false;
		updateEditorAfterURLChange();
	});
}

function updateEditorAfterURLChange(): void {
	// TODO restore serialized view state.
	const sel = getSelectionFromRouter();
	if (!sel) {
		return;
	}

	const editor = getEditorInstance();
	editor.setSelection(sel);
	editor.revealRangeInCenter(sel);
}

// registerEditorCallbacks attaches custom Sourcegraph handling to the workbench editor lifecycle.
export function registerEditorCallbacks(): void {
	const editorService = Services.get(ICodeEditorService) as ICodeEditorService;
	editorService.onCodeEditorAdd(updateEditor);
}

// editorOpened is called whenever a new editor is created or activated. E.g:
//  - on page load
//  - from file explorer
//  - for a cross-file j2d
function editorOpened(event: IModelChangedEvent): void {
	if (forceSyncInProgress) {
		return;
	}
	const resource = event.newModelUrl;
	let {repo, rev, path} = URIUtils.repoParams(resource);
	if (rev === "HEAD") {
		rev = null;
	}

	const workspaceService = Services.get(IWorkspaceContextService) as IWorkspaceContextService;
	workspaceService.setWorkspace({ resource: resource.with({ fragment: "" }) });

	router.push(urlToBlob(repo, rev, path));
}

function updateEditor(editor: ICodeEditor): void {
	updateEditorInstance(editor);

	// Listeners
	editor.onDidChangeModel(editorOpened);
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
