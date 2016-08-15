// tslint:disable: typedef ordered-imports

import * as React from "react";
import {Link} from "react-router";
import {Container} from "sourcegraph/Container";
import * as Dispatcher from "sourcegraph/Dispatcher";
import trimLeft from "lodash.trimleft";
import {TreeStore} from "sourcegraph/tree/TreeStore";
import "sourcegraph/tree/TreeBackend";
import * as TreeActions from "sourcegraph/tree/TreeActions";
import {Header} from "sourcegraph/components/Header";
import {urlToBlob} from "sourcegraph/blob/routes";
import {urlToTree} from "sourcegraph/tree/routes";
import {httpStatusCode} from "sourcegraph/util/httpStatusCode";
import * as classNames from "classnames";

import {FileIcon, FolderIcon} from "sourcegraph/components/Icons";

import * as styles from "./styles/Tree.css";

const EMPTY_PATH = [];

function pathSplit(path: string): string[] {
	if (path === "") {
		throw new Error("invalid empty path");
	}
	if (path === "/") {
		return EMPTY_PATH;
	}
	path = trimLeft(path, "/");
	return path.split("/");
}

function pathJoin(pathComponents: string[]): string {
	if (pathComponents.length === 0) {
		return "/";
	}
	return pathComponents.join("/");
}

function pathJoin2(a: string, b: string): string {
	if (!a || a === "/") {
		return b;
	}
	return `${a}/${b}`;
}

function pathDir(path: string): string {
	// Remove last item from path.
	const parts = pathSplit(path);
	return pathJoin(parts.splice(0, parts.length - 1));
}

interface Props {
	repo: string;
	rev: string | null;
	commitID: string;
	path: string;
	location: any;
	route?: ReactRouter.Route;
}

type State = {
	// prop types
	repo: string;
	rev: string | null;
	commitID?: string;
	path?: string;
	location?: any;
	route?: ReactRouter.Route;

	// other state fields
	fileResults: any; // Array<any> | {Error: any};
	fileTree?: any;
}

export class TreeList extends Container<Props, State> {
	static contextTypes = {
		router: React.PropTypes.object.isRequired,
		user: React.PropTypes.object,
	};

	constructor(props: Props) {
		super(props);
		this.state = {
			repo: "",
			rev: null,
			fileResults: [],
		};
	}

	stores(): FluxUtils.Store<any>[] { return [TreeStore]; }

	reconcileState(state: State, props: Props): void {
		let prevPath = state.path;
		Object.assign(state, props);

		let newFileTree = TreeStore.fileTree.get(state.repo, state.commitID);
		if (newFileTree !== state.fileTree || prevPath !== state.path) {
			state.fileTree = newFileTree;
			state.fileResults = [];

			if (state.fileTree && state.path) {
				let dirLevel = state.fileTree;
				let err;
				for (const part of pathSplit(state.path)) {
					let dirKey = `!${part}`; // dirKey is prefixed to avoid clash with predefined fields like "constructor"
					if (dirLevel.Dirs[dirKey]) {
						dirLevel = dirLevel.Dirs[dirKey];
					} else {
						if (!dirLevel.Dirs[dirKey] && !dirLevel.Files[part]) {
							err = {response: {status: 404}};
						}
						break;
					}
				}

				const pathPrefix = state.path.replace(/^\/$/, "");
				const dirs = !err ? Object.keys(dirLevel.Dirs).map(dirKey => ({
					name: dirKey.substr(1), // dirKey is prefixed to avoid clash with predefined fields like "constructor"
					isDirectory: true,
					isParentDirectory: false,
					path: pathJoin2(pathPrefix, dirKey.substr(1)),
					url: urlToTree(state.repo, state.rev, pathJoin2(pathPrefix, dirKey.substr(1))),
				})) : [];
				// Add parent dir link if showing a subdir.
				if (pathPrefix) {
					const parentDir = pathDir(pathPrefix);
					dirs.unshift({
						name: "..",
						isDirectory: true,
						isParentDirectory: true,
						path: parentDir,
						url: urlToTree(state.repo, state.rev, parentDir),
					});
				}

				const files = !err ? dirLevel.Files.map(file => ({
					name: file,
					isDirectory: false,
					url: urlToBlob(state.repo, state.rev, pathJoin2(pathPrefix, file)),
				})) : [];
				// TODO Handle errors in a more standard way.
				state.fileResults = !err ? dirs.concat(files) : {Error: err};
			}
		}
	}

	onStateTransition(prevState: State, nextState: State): void {
		if ((nextState.repo !== prevState.repo || nextState.commitID !== prevState.commitID) && nextState.commitID) {
			Dispatcher.Backends.dispatch(new TreeActions.WantFileList(nextState.repo, nextState.commitID));
		}
	}

	_listItems(): Array<any> {
		const items = this.state.fileResults;
		const emptyItem = <div className={classNames(styles.list_item, styles.list_item_empty)} key="_nofiles"><i>No matches.</i></div>;
		if (!items || items.length === 0) {
			return [emptyItem];
		}

		let list: any[] = [];
		for (let i = 0; i < items.length; i++) {
			let item = items[i];
			let itemURL = item.url;

			let icon;
			if (item.isParentDirectory) {
				icon = null;
			} else if (item.isDirectory) {
				icon = <FolderIcon className={styles.icon} />;
			} else {
				icon = <FileIcon className={styles.icon} />;
			}

			let key = `f:${itemURL}`;
			list.push(
				<Link className={classNames(styles.list_item, item.isParentDirectory && styles.parent_dir)}
					to={itemURL}
					key={key}>
					{icon}
					{item.name}
				</Link>
			);
		}

		return list;
	}

	render(): JSX.Element | null {
		if (this.state.fileResults && this.state.fileResults.Error) {
			let code = httpStatusCode(this.state.fileResults.Error);
			return (
				<Header
					title={`${code}`}
					subtitle={code === 404 ? `Directory not found.` : "Directory is not available."} />
			);
		}

		let listItems = this._listItems() || [];
		return (
			<div className={styles.tree_common}>
				<div className={styles.list_header}>
					Files
				</div>
				<div className={styles.list_item_group}>
					{listItems}
				</div>
			</div>
		);
	}
}
