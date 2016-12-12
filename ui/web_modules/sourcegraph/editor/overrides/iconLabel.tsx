import * as drop from "lodash/drop";
import { FileLabel } from "vs/base/browser/ui/iconLabel/iconLabel";
import { IWorkspaceProvider } from "vs/base/common/labels";
import uri from "vs/base/common/uri";

FileLabel.prototype.setFile = function (file: uri, provider: IWorkspaceProvider): void {
	const path = file.path + "/" + file.fragment;
	const dirs = drop(path.split("/"));
	const base = dirs.pop();
	this.setValue(base, dirs.join("/"));
};
