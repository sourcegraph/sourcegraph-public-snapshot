import { browserHistory } from "react-router";
import URI from "vs/base/common/uri";
import { IEditorInput } from "vs/platform/editor/common/editor";
import { IInstantiationService } from "vs/platform/instantiation/common/instantiation";
import { EditorPart } from "vs/workbench/browser/parts/editor/editorPart";

import { urlToBlob } from "sourcegraph/blob/routes";
import { URIUtils } from "sourcegraph/core/uri";

export function configureEditor(editor: EditorPart, resource: URI, instantiationService: IInstantiationService): void {
	const stacks = editor.getStacksModel();
	stacks.activeGroup.onEditorActivated(editorOpened);
}

function editorOpened(input: IEditorInput): void {
	if (!global.window) {
		return;
	}
	let resource;
	if (input["resource"]) {
		resource = (input as any).resource;
	} else {
		throw "Couldn't find resource";
	}
	// TODO set workspace on workspace jump.
	const {repo, rev, path} = URIUtils.repoParams(resource);
	browserHistory.push(urlToBlob(repo, rev, path));
}
