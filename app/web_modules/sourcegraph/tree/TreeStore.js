import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as TreeActions from "sourcegraph/tree/TreeActions";

function keyFor(repo, rev, path) {
	return `${repo}#${rev}#${path || ""}`;
}

export class TreeStore extends Store {
	reset() {
		this.commits = deepFreeze({
			content: {},
			get(repo, rev, path) {
				return this.content[keyFor(repo, rev, path)] || null;
			},
		});
		this.fileLists = deepFreeze({
			content: {},
			get(repo, rev) {
				return this.content[keyFor(repo, rev)] || null;
			},
		});
		this.fileTree = deepFreeze({
			content: {},
			get(repo, rev) {
				return this.content[keyFor(repo, rev)] || null;
			},
		});
		this.srclibDataVersions = deepFreeze({
			content: {},
			get(repo, rev, path) {
				return this.content[keyFor(repo, rev, path)] || null;
			},
		});
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
				let fileTree = {Dirs: {}, Files: []};
				action.fileList.Files.forEach(file => {
					const parts = file.split("/");
					let node = fileTree;
					parts.forEach((part, i) => {
						if (i === parts.length - 1) {
							node.Files.push(part);
						} else if (!node.Dirs[part]) {
							node.Dirs[part] = {Dirs: {}, Files: []};
						}
						node = node.Dirs[part];
					});
				});
				this.fileLists = deepFreeze(Object.assign({}, this.fileLists, {
					content: Object.assign({}, this.fileLists.content, {
						[keyFor(action.repo, action.rev)]: action.fileList,
					}),
				}));
				this.fileTree = deepFreeze(Object.assign({}, this.fileTree, {
					content: Object.assign({}, this.fileTree.content, {
						[keyFor(action.repo, action.rev)]: fileTree,
					}),
				}));
				break;
			}

		case TreeActions.FetchedSrclibDataVersion:
			this.srclibDataVersions = deepFreeze(Object.assign({}, this.srclibDataVersions, {
				content: Object.assign({}, this.srclibDataVersions.content, {
					[keyFor(action.repo, action.rev, action.path)]: action.version,
				}),
			}));
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new TreeStore(Dispatcher);
