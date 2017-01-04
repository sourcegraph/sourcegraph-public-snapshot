import * as throttle from "lodash/throttle";
import { ICodeEditor } from "vs/editor/browser/editorBrowser";
import { EmbeddedCodeEditorWidget } from "vs/editor/browser/widget/embeddedCodeEditorWidget";
import { ICursorSelectionChangedEvent } from "vs/editor/common/editorCommon";
import { ICodeEditorService } from "vs/editor/common/services/codeEditorService";
import { IWorkspaceContextService } from "vs/platform/workspace/common/workspace";
import { EditorInput } from "vs/workbench/common/editor";
import { EditorStacksModel } from "vs/workbench/common/editor/editorStacksModel";
import { IWorkbenchEditorService } from "vs/workbench/services/editor/common/editorService";
import { IEditorGroupService } from "vs/workbench/services/group/common/groupService";

import { getBlobPropsFromRouter, getSelectionFromRouter, router } from "sourcegraph/app/router";
import { urlToBlob } from "sourcegraph/blob/routes";
import { URIUtils } from "sourcegraph/core/uri";
import { getEditorInstance, updateEditorInstance } from "sourcegraph/editor/Editor";
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
	const editorService = Services.get(ICodeEditorService) as ICodeEditorService;
	editorService.onCodeEditorAdd(updateEditor);

	// A group is a set of editors. It is used in VSCode to display the left,
	// right and center tab groups. A group has a stack of editors. We use the
	// stack to determine which file is currently focused.
	const groupService = Services.get(IEditorGroupService) as IEditorGroupService;
	const stacks = groupService.getStacksModel();
	if (!(stacks instanceof EditorStacksModel)) {
		throw "Expected IEditorStacksModel to have concrete type EditorStacksModel";
	}
	stacks.onGroupActivated((group) => {
		group.onEditorActivated(editorOpened);
	});
}

// editorOpened is called whenever a new editor is created or activated. E.g:
//  - on page load
//  - from file explorer
//  - for a cross-file j2d
function editorOpened(input: EditorInput): void {
	if (forceSyncInProgress) {
		return;
	}
	if (!input["resource"]) {
		throw "Expected input to have resource attribute.";
	}
	const resource = input["resource"];
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
