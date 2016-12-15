import * as orig from "vscode/src/vs/workbench/browser/labels";

import { IWorkspaceProvider } from "vs/base/common/labels";
import URI from "vs/base/common/uri";

import { setFile } from "sourcegraph/workbench/overrides/iconLabel";

export class FileLabel extends orig.FileLabel {
	setFile(file: URI, provider: IWorkspaceProvider): void {
		setFile(file, provider);
	}
}

export const ResourceLabel = orig.ResourceLabel;
export const EditorLabel = orig.EditorLabel;
export const getIconClasses = orig.getIconClasses;
