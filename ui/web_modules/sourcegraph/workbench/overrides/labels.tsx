import * as vs from "vscode/src/vs/workbench/browser/labels";

import { IWorkspaceProvider } from "vs/base/common/labels";
import URI from "vs/base/common/uri";

import { setFile } from "sourcegraph/workbench/overrides/iconLabel";

export class FileLabel extends vs.FileLabel {
	setFile(file: URI, provider: IWorkspaceProvider): void {
		setFile(file, provider);
	}
}

export const ResourceLabel = vs.ResourceLabel;
export const EditorLabel = vs.EditorLabel;
export const getIconClasses = vs.getIconClasses;
