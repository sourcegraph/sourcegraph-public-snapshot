import * as vs from "vscode/src/vs/base/browser/ui/iconLabel/iconLabel";

import * as drop from "lodash/drop";
import { IWorkspaceProvider } from "vs/base/common/labels";
import URI from "vs/base/common/uri";

// We override the file label because VSCode uses different URI conventions
// than we do. This is required to make the references view file list have
// reasonable names.

export class FileLabel extends vs.FileLabel {
	setFile(file: URI, provider: IWorkspaceProvider): void {
		setFile(file, provider);
	}
}

export function setFile(file: URI, provider: IWorkspaceProvider): void {
	const path = file.path + "/" + file.fragment;
	const dirs = drop(path.split("/"));
	const base = dirs.pop();
	this.setValue(base, dirs.join("/"));
};

export const IconLabel = vs.IconLabel;
