import URI from "vs/base/common/uri";
import { IInstantiationService } from "vs/platform/instantiation/common/instantiation";
import { EditorPart } from "vs/workbench/browser/parts/editor/editorPart";
import { IEditorIdentifier } from "vs/workbench/common/editor";

import { urlToBlob } from "sourcegraph/blob/routes";
import { URIUtils } from "sourcegraph/core/uri";

export function configureEditor(editor: EditorPart, resource: URI, instantiationService: IInstantiationService): void {
	editor.getStacksModel().onEditorOpened(editorOpened);
}

function editorOpened(editorID: IEditorIdentifier): void {
	if (global.window) {
		let resource;
		if (editorID.editor["resource"]) {
			resource = (editorID.editor as any).resource;
		} else {
			throw "Couldn't find resource";
		}
		const {repo, rev, path} = URIUtils.repoParams(resource);
		history.pushState({}, document.title, urlToBlob(repo, rev, path));
	}
}
