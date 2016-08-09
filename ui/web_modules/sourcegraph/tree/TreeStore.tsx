// tslint:disable

import {Store} from "sourcegraph/Store";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {deepFreeze} from "sourcegraph/util/deepFreeze";
import * as TreeActions from "sourcegraph/tree/TreeActions";
import "sourcegraph/tree/TreeBackend";

function keyFor(repo, rev, path?) {
	return `${repo}#${rev}#${path || ""}`;
}

class TreeStoreClass extends Store<any> {
	commits: any;
	fileLists: any;
	fileTree: any;
	srclibDataVersions: any;

	reset() {
		this.commits = deepFreeze({
			content: {},
			get(repo, rev, path) {
				return this.content[keyFor(repo, rev, path)] || null;
			},
		});
		this.fileLists = deepFreeze({
			content: {},
			get(repo, commitID) {
				return this.content[keyFor(repo, commitID)] || null;
			},
		});
		this.fileTree = deepFreeze({
			content: {},
			get(repo, commitID) {
				return this.content[keyFor(repo, commitID)] || null;
			},
		});
		this.srclibDataVersions = deepFreeze({
			content: {},
			get(repo, commitID, path) {
				return this.content[keyFor(repo, commitID, path)] || null;
			},
		});
	}

	toJSON(): any {
		return {
			commits: this.commits,
			fileLists: this.fileLists,
			fileTree: this.fileTree,
			srclibDataVersions: this.srclibDataVersions,
		};
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case TreeActions.CommitFetched:
			this.commits = deepFreeze(Object.assign({}, this.commits, {
				content: Object.assign({}, this.commits.content, {
					[keyFor(action.repo, action.rev, action.path)]: action.commit,
				}),
			}));
			break;

		case TreeActions.FileListFetched:
			{
				let fileTree = {Dirs: {}, Files: [] as (any[])};
				if (action.fileList && action.fileList.Files) {
					action.fileList.Files.forEach(file => {
						const parts = file.split("/");
						let node = fileTree;
						parts.forEach((part, i) => {
							let dirKey = `!${part}`; // dirKey is prefixed to avoid clash with predefined fields like "constructor"
							if (i === parts.length - 1) {
								node.Files.push(part);
							} else if (!node.Dirs[dirKey]) {
								node.Dirs[dirKey] = {Dirs: {}, Files: []};
							}
							node = node.Dirs[dirKey];
						});
					});
				}
				this.fileLists = deepFreeze(Object.assign({}, this.fileLists, {
					content: Object.assign({}, this.fileLists.content, {
						[keyFor(action.repo, action.commitID)]: action.fileList,
					}),
				}));
				this.fileTree = deepFreeze(Object.assign({}, this.fileTree, {
					content: Object.assign({}, this.fileTree.content, {
						[keyFor(action.repo, action.commitID)]: fileTree,
					}),
				}));
				break;
			}

		case TreeActions.FetchedSrclibDataVersion:
			this.srclibDataVersions = deepFreeze(Object.assign({}, this.srclibDataVersions, {
				content: Object.assign({}, this.srclibDataVersions.content, {
					[keyFor(action.repo, action.commitID, action.path)]: action.version,
				}),
			}));
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export const TreeStore = new TreeStoreClass(Dispatcher.Stores);
