import * as throttle from "lodash/throttle";
import URI from "vs/base/common/uri";
import { ICodeEditor } from "vs/editor/browser/editorBrowser";
import { EmbeddedCodeEditorWidget } from "vs/editor/browser/widget/embeddedCodeEditorWidget";
import { ICursorSelectionChangedEvent } from "vs/editor/common/editorCommon";
import { ICodeEditorService } from "vs/editor/common/services/codeEditorService";
import { IEditorService } from "vs/platform/editor/common/editor";
import { IWorkspaceContextService } from "vs/platform/workspace/common/workspace";
import { IWorkbenchEditorService } from "vs/workbench/services/editor/common/editorService";

import { getBlobPropsFromRouter, getSelectionFromRouter, router } from "sourcegraph/app/router";
import { urlToBlob } from "sourcegraph/blob/routes";
import { URIUtils } from "sourcegraph/core/uri";
import { getEditorInstance, updateEditorInstance } from "sourcegraph/editor/Editor";
import { WorkbenchEditorService } from "sourcegraph/workbench/overrides/editorService";
import { Services } from "sourcegraph/workbench/services";

// forceSyncInProgress is a mutex. We only want to open the editor to some
// input if it has not already been done. If the change is a result of
// syncEditorWithRouter, then we don't need to run editorOpened because the URL
// is already up to date. This is necessary because the two functions are
// cyclic, and we only want to run them once for each action.
let forceSyncInProgress: boolean;

// syncEditorWithURL forces the editor model to match current URL blob properties.
export function syncEditorWithRouter(): void {
	const {repo, rev, path} = getBlobPropsFromRouter();
	const resource = URIUtils.pathInRepo(repo, rev, path);
	const editorService = Services.get(IWorkbenchEditorService) as IWorkbenchEditorService;
	const workspaceService = Services.get(IWorkspaceContextService) as IWorkspaceContextService;
	workspaceService.setWorkspace({ resource: resource.with({ fragment: "" }) });
	if (forceSyncInProgress) {
		return;
	}
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
	const codeEditorService = Services.get(ICodeEditorService) as ICodeEditorService;
	codeEditorService.onCodeEditorAdd(updateEditor);

	const editorService = Services.get(IEditorService) as WorkbenchEditorService;
	editorService.onDidOpenEditor(editorOpened);
}

// editorOpened is called whenever a new editor is created or activated. E.g:
//  - on page load
//  - from file explorer
//  - for a cross-file j2d
function editorOpened(resource: URI): void {
	if (forceSyncInProgress) {
		return;
	}
	let {repo, rev, path} = URIUtils.repoParams(resource);
	if (rev === "HEAD") {
		rev = null;
	}

	const workspaceService = Services.get(IWorkspaceContextService) as IWorkspaceContextService;
	workspaceService.setWorkspace({ resource: resource.with({ fragment: "" }) });

	forceSyncInProgress = true;
	router.push(urlToBlob(repo, rev, path));
	forceSyncInProgress = false;
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
