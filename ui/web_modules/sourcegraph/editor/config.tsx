import URI from "vs/base/common/uri";
import { IEditorInput } from "vs/platform/editor/common/editor";
import { ServiceCollection } from "vs/platform/instantiation/common/serviceCollection";
import { EditorPart } from "vs/workbench/browser/parts/editor/editorPart";
import { IWorkbenchEditorService } from "vs/workbench/services/editor/common/editorService";

import { urlToBlob } from "sourcegraph/blob/routes";
import { URIUtils } from "sourcegraph/core/uri";

export function configureEditor(editor: EditorPart, resource: URI): void {
	const stacks = editor.getStacksModel();
	stacks.activeGroup.onEditorActivated(editorOpened);
}

function editorOpened(input: IEditorInput): void {
	if (!global.window || updating) {
		return;
	}
	let resource;
	if (input["resource"]) {
		resource = (input as any).resource;
	} else {
		throw "Couldn't find resource.";
	}
	// TODO set workspace on workspace jump.
	const {repo, rev, path} = URIUtils.repoParams(resource);
	history.pushState({}, "", urlToBlob(repo, rev, path));
}

let updating = false;
export function updateEditor(editor: EditorPart, resource: URI, services: ServiceCollection): void {
	const editorService = services.get(IWorkbenchEditorService) as IWorkbenchEditorService;
	updating = true;
	editorService.openEditor({ resource });
	updating = false;
}
