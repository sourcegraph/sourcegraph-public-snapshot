import * as React from "react";
import ReactDOM from "react-dom";
import {Link} from "react-router";
import fuzzysearch from "fuzzysearch";
import fuzzy_score from "sourcegraph/tree/fuzzy_score";
import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import debounce from "lodash.debounce";
import trimLeft from "lodash.trimleft";
import TreeStore from "sourcegraph/tree/TreeStore";
import "sourcegraph/tree/TreeBackend";
import * as TreeActions from "sourcegraph/tree/TreeActions";
import Header from "sourcegraph/components/Header";
import {urlToBlob} from "sourcegraph/blob/routes";
import {urlToTree} from "sourcegraph/tree/routes";
import httpStatusCode from "sourcegraph/util/httpStatusCode";
import type {Route} from "react-router";

import {Input} from "sourcegraph/components";
import {FileIcon, FolderIcon} from "sourcegraph/components/Icons";

import CSSModules from "react-css-modules";
import styles from "./styles/Tree.css";

const FILE_LIMIT = 15;
const EMPTY_PATH = [];

function pathSplit(path: string): string[] {
	if (path === "") throw new Error("invalid empty path");
	if (path === "/") return EMPTY_PATH;
	path = trimLeft(path, "/");
	return path.split("/");
}

function pathJoin(pathComponents: string[]): string {
	if (pathComponents.length === 0) return "/";
	return pathComponents.join("/");
}

function pathJoin2(a: string, b: string): string {
	if (!a || a === "/") return b;
	return `${a}/${b}`;
}

function pathDir(path: string): string {
	// Remove last item from path.
	const parts = pathSplit(path);
	return pathJoin(parts.splice(0, parts.length - 1));
}

type TreeSearchProps = {
	repo: string;
	rev: ?string;
	commitID: string;
	path: string;
	overlay: boolean;
	prefetch: ?boolean;
	location: Location;
	route: Route;
	onChangeQuery: (query: string) => void;
	onSelectPath: (path: string) => void;
}

type TreeSearchState = {
	// prop types
	repo: string;
	rev?: string;
	commitID?: string;
	path?: string;
	overlay?: boolean;
	prefetch?: boolean;
	location?: Location;
	route?: Route;
	onChangeQuery: (query: string) => void;
	onSelectPath: (path: string) => void;

	// other state fields
	query: string;
	focused: boolean;
	selectionIndex: number;
	fileResults: any; // Array<any> | {Error: any};
	srclibDataVersion?: Object;
	fileTree?: any;
	fileList?: any;
	fuzzyFinder?: any;
}


class TreeSearch extends Container {
	static propTypes = {
		repo: React.PropTypes.string.isRequired,
		rev: React.PropTypes.string,
		commitID: React.PropTypes.string.isRequired,
		path: React.PropTypes.string.isRequired,
		onSelectPath: React.PropTypes.func.isRequired,
		onChangeQuery: React.PropTypes.func.isRequired,
		query: React.PropTypes.string,
		overlay: React.PropTypes.bool,
		prefetch: React.PropTypes.bool,
		location: React.PropTypes.object,
		route: React.PropTypes.object,
	};

	props: TreeSearchProps;
	state: TreeSearchState;

	static contextTypes = {
		router: React.PropTypes.object.isRequired,
		user: React.PropTypes.object,
	};

	constructor(props: TreeSearchProps) {
		super(props);
		this.state = {
			repo: "",
			query: "",
			focused: !props.overlay,
			selectionIndex: 0,
			onChangeQuery: (query: string) => { /* do nothing */ },
			onSelectPath: (path: string) => { /* do nothing */ },
			fileResults: [],
		};
		this._queryInput = null;
		this._handleKeyDown = this._handleKeyDown.bind(this);
		this._scrollToVisibleSelection = this._scrollToVisibleSelection.bind(this);
		this._setSelectedItem = this._setSelectedItem.bind(this);
		this._handleInput = this._handleInput.bind(this);
		this._onSelection = debounce(this._onSelection.bind(this), 100, {leading: false, trailing: true}); // Prevent rapid repeated selections
		this._debouncedSetQuery = debounce((query) => {
			if (query !== this.state.query) {
				setTimeout(() => this.props.onChangeQuery(query));
			}
		}, 25, {leading: false, trailing: true});
	}

	componentDidMount() {
		super.componentDidMount();
		if (global.document) {
			document.addEventListener("keydown", this._handleKeyDown);
		}
	}

	componentWillUnmount() {
		super.componentWillUnmount();
		if (global.document) {
			document.removeEventListener("keydown", this._handleKeyDown);
		}
	}

	stores(): Array<Object> { return [TreeStore]; }

	reconcileState(state: TreeSearchState, props: TreeSearchProps): void {
		Object.assign(state, props);

		state.query = props.query || "";

		state.fileTree = TreeStore.fileTree.get(state.repo, state.commitID);
		state.fileList = TreeStore.fileLists.get(state.repo, state.commitID);
	}

	onStateTransition(prevState: TreeSearchState, nextState: TreeSearchState) {
		const prefetch = nextState.prefetch && nextState.prefetch !== prevState.prefetch;
		if (prefetch || nextState.repo !== prevState.repo || nextState.commitID !== prevState.commitID) {
			if (nextState.commitID) {
				Dispatcher.Backends.dispatch(new TreeActions.WantFileList(nextState.repo, nextState.commitID));
			}
		}

		if (prevState.path !== nextState.path || prevState.query !== nextState.query || prevState.fileList !== nextState.fileList || prevState.fileTree !== nextState.fileTree) {
			nextState.fileResults = [];
			nextState.selectionIndex = 0;

			// Show entire file tree as file results.
			//
			// TODO Find a better way to do this without updating state in onStateTransition.
			if (!nextState.query) {
				if (nextState.fileTree) {
					let dirLevel = nextState.fileTree;
					let err;
					for (const part of pathSplit(nextState.path)) {
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

					const pathPrefix = nextState.path.replace(/^\/$/, "");
					const dirs = !err ? Object.keys(dirLevel.Dirs).map(dirKey => ({
						name: dirKey.substr(1), // dirKey is prefixed to avoid clash with predefined fields like "constructor"
						isDirectory: true,
						path: pathJoin2(pathPrefix, dirKey.substr(1)),
						url: urlToTree(nextState.repo, nextState.rev, pathJoin2(pathPrefix, dirKey.substr(1))),
					})) : [];
					// Add parent dir link if showing a subdir.
					if (pathPrefix) {
						const parentDir = pathDir(pathPrefix);
						dirs.unshift({
							name: "..",
							isDirectory: true,
							isParentDirectory: true,
							path: parentDir,
							url: urlToTree(nextState.repo, nextState.rev, parentDir),
						});
					}

					const files = !err ? dirLevel.Files.map(file => ({
						name: file,
						isDirectory: false,
						url: urlToBlob(nextState.repo, nextState.rev, pathJoin2(pathPrefix, file)),
					})) : [];
					// TODO Handle errors in a more standard way.
					nextState.fileResults = !err ? dirs.concat(files) : {Error: err};
				}

			} else {
				const query = nextState.query.toLowerCase();
				if (nextState.fileList && nextState.fileList.Files) {
					nextState.fileResults = nextState.fileList.Files
						.map(f => fuzzysearch(query, f.toLowerCase()) && fuzzy_score(nextState.query, f))
						.filter(f => f)               // drop partial matches
						.sort((a, b) => b[0] - a[0])  // descending order by score
						.map(f => ({
							name: f[1],
							isDirectory: false,
							url: urlToBlob(nextState.repo, nextState.rev, f[1]),
						}));
				}
			}
		}
	}

	_navigateTo(url: string) {
		this.context.router.push(url);
	}

	_handleKeyDown(e: KeyboardEvent) {
		if (!this.state.focused) return;

		let idx, max;
		switch (e.keyCode) {
		case 40: // ArrowDown
			idx = this._normalizedSelectionIndex();
			max = this._numResults();

			this.setState({
				selectionIndex: idx + 1 >= max ? 0 : idx + 1,
			}, this._scrollToVisibleSelection);

			this._temporarilyIgnoreMouseSelection();
			e.preventDefault();
			break;

		case 38: // ArrowUp
			idx = this._normalizedSelectionIndex();
			max = this._numResults();

			this.setState({
				selectionIndex: idx < 1 ? max-1 : idx-1,
			}, this._scrollToVisibleSelection);

			this._temporarilyIgnoreMouseSelection();
			e.preventDefault();
			break;

		case 13: // Enter
			this._onSelection();
			this._temporarilyIgnoreMouseSelection();
			e.preventDefault();
			break;

		default:
			// Changes to the input value are handled by _handleInput.
			break;
		}
	}

	_handleInput(e: KeyboardEvent) {
		if (this.state.focused) {
			this._debouncedSetQuery(this._queryInput ? this._queryInput.value : "");
		}
	}

	_scrollToVisibleSelection() {
		if (this._selectedItem) ReactDOM.findDOMNode(this._selectedItem).scrollIntoView(false);
	}

	_setSelectedItem(e: any) {
		this._selectedItem = e;
	}

	_numFileResults(): number {
		if (!this.state.fileResults) return 0;
		let numFileResults = Math.min(this.state.fileResults.length, FILE_LIMIT);
		// Override file results to show full directory tree on empty query, or when
		// query is 3+ chars.
		if (!this.state.query || this.state.query.length >= 3) numFileResults = this.state.fileResults.length;
		return numFileResults;
	}

	_numResults(): number {
		return this._numFileResults();
	}

	_normalizedSelectionIndex(): number {
		return Math.min(this.state.selectionIndex, this._numResults() - 1);
	}

	_onSelection() {
		const i = this._normalizedSelectionIndex();
		if (!this.state.fileResults) return;
		let result = this.state.fileResults[i];
		if (!result) return;
		if (result.isDirectory) {
			this.state.onSelectPath(result.path);
		} else {
			this._navigateTo(result.url);
		}
	}

	// returns the selected directory name, or null
	_getSelectedPathPart(): ?string {
		const i = this._normalizedSelectionIndex();
		const result = this.state.fileResults[i];
		if (result.isDirectory) return result.name;
		return null;
	}

	// returns the selected file name, or null
	_getSelectedFile(): ?string {
		const i = this._normalizedSelectionIndex();
		const result = this.state.fileResults[i];
		if (!result.isDirectory) return result.name;
		return null;
	}

	_listItems(): Array<any> {
		const items = this.state.fileResults;
		const emptyItem = <div styleName="list-item list-item-empty" key="_nofiles"><i>No matches.</i></div>;
		if (!items || items.length === 0) return [emptyItem];

		let list = [],
			limit = items.length > FILE_LIMIT ? FILE_LIMIT : items.length;

		// Override limit if query is empty to show the full directory tree.
		if (!this.state.query) limit = items.length;

		for (let i = 0; i < limit; i++) {
			let item = items[i],
				itemURL = item.url;

			const selected = this._normalizedSelectionIndex() === i;

			let icon;
			if (item.isParentDirectory) icon = null;
			else if (item.isDirectory) icon = <FolderIcon styleName="icon" />;
			else icon = <FileIcon styleName="icon" />;

			let key = `f:${itemURL}`;
			list.push(
				<Link styleName={`${selected ? "list-item-selected" : "list-item"} ${item.isParentDirectory ? "parent-dir" : ""}`}
					onMouseOver={(ev) => this._mouseSelectItem(ev, i)}
					ref={selected ? this._setSelectedItem : null}
					to={itemURL}
					key={key}>
					{icon}
					{item.name}
				</Link>
			);
		}

		return list;
	}

	_selectItem(i: number): void {
		this.setState({
			selectionIndex: i,
		});
	}

	// _mouseSelectItem causes i to be selected ONLY IF the user is using the
	// mouse to select. It ignores the case where the user is using the up/down
	// keys to change the selection and the window scrolls, causing the mouse cursor
	// to incidentally hover a different element. We ignore mouse selections except
	// those where the mouse was actually moved.
	_mouseSelectItem(ev: MouseEvent, i: number): void {
		if (this._ignoreMouseSelection) return;
		this._selectItem(i);
	}

	// _temporarilyIgnoreMouseSelection is used to ignore mouse selections. See
	// _mouseSelectItem.
	_temporarilyIgnoreMouseSelection() {
		if (!this._debouncedUnignoreMouseSelection) {
			this._debouncedUnignoreMouseSelection = debounce(() => {
				this._ignoreMouseSelection = false;
			}, 200, {leading: false, trailing: true});
		}
		this._debouncedUnignoreMouseSelection();
		this._ignoreMouseSelection = true;
	}

	render() {
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
			<div styleName="tree-common">
				<div styleName="input-container">
					<Input type="text"
						id="search-input"
						block={true}
						onFocus={() => this.setState({focused: true})}
						onBlur={(e) => this.setState({focused: false})}
						onInput={this._handleInput}
						autoFocus={true}
						defaultValue={this.state.query}
						placeholder="Jump to files..."
						spellCheck={false}
						domRef={(e) => this._queryInput = e} />
				</div>

				<div styleName="list-header">
					Files
				</div>
				<div styleName="list-item-group">
					{listItems}
				</div>
			</div>
		);
	}
}

export default CSSModules(TreeSearch, styles, {allowMultiple: true});
