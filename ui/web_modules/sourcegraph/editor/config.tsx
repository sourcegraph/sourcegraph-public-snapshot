import URI from "vs/base/common/uri";
import { IEditorInput } from "vs/platform/editor/common/editor";
import { ServiceCollection } from "vs/platform/instantiation/common/serviceCollection";
import { EditorPart } from "vs/workbench/browser/parts/editor/editorPart";
import { IWorkbenchEditorService } from "vs/workbench/services/editor/common/editorService";

import { parseBlobURL, urlToBlob } from "sourcegraph/blob/routes";
import { URIUtils } from "sourcegraph/core/uri";
import { getResource } from "sourcegraph/workbench/utils";

export function configureEditor(editor: EditorPart, resource: URI): void {
	const stacks = editor.getStacksModel();
	stacks.activeGroup.onEditorActivated(editorOpened);
}

// editorOpened is called whenever a new editor is created or activated. When
// this event happens, we update the URL to match the editor file.
function editorOpened(input: IEditorInput): void {
	if (!global.window) {
		return;
	}
	const resource = getResource(input);
	// TODO set workspace on workspace jump.
	const oldParams = parseBlobURL(document.location.toString());
	const currentURL = urlToBlob(oldParams.repo, oldParams.rev, oldParams.path);
	let {repo, rev, path} = URIUtils.repoParams(resource);
	if (rev === "HEAD") {
		rev = null;
	}
	const url = urlToBlob(repo, rev, path);
	if (url === currentURL) {
		return;
	}
	history.pushState({}, "", url);
}

export function updateEditor(editor: EditorPart, resource: URI, services: ServiceCollection): void {
	const editorService = services.get(IWorkbenchEditorService) as IWorkbenchEditorService;
	editorService.openEditor({ resource });
}
