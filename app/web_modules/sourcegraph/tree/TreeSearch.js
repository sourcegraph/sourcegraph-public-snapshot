// @flow

import React from "react";
import ReactDOM from "react-dom";
import Fuze from "fuse.js";
import {Link} from "react-router";
import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import debounce from "lodash/function/debounce";
import trimLeft from "lodash/string/trimLeft";
import TreeStore from "sourcegraph/tree/TreeStore";
import DefStore from "sourcegraph/def/DefStore";
import "sourcegraph/tree/TreeBackend";
import "sourcegraph/def/DefBackend";
import * as TreeActions from "sourcegraph/tree/TreeActions";
import * as DefActions from "sourcegraph/def/DefActions";
import {qualifiedNameAndType} from "sourcegraph/def/Formatter";
import {urlToBlob} from "sourcegraph/blob/routes";
import {urlToDef} from "sourcegraph/def/routes";
import {urlToTree} from "sourcegraph/tree/routes";
import {urlToBuilds} from "sourcegraph/build/routes";
import type {Def} from "sourcegraph/def";
import type {Route} from "react-router";

import {Input, Loader, Icon, RepoLink} from "sourcegraph/components";

import breadcrumb from "sourcegraph/util/breadcrumb";

import CSSModules from "react-css-modules";
import styles from "./styles/Tree.css";

const SYMBOL_LIMIT = 5;
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

class TreeSearch extends Container {
	static propTypes = {
		repo: React.PropTypes.string.isRequired,
		rev: React.PropTypes.string.isRequired,
		path: React.PropTypes.string.isRequired,
		onSelectPath: React.PropTypes.func.isRequired,
		onChangeQuery: React.PropTypes.func.isRequired,
		query: React.PropTypes.string,
		overlay: React.PropTypes.bool,
		prefetch: React.PropTypes.bool,
		location: React.PropTypes.object,
		route: React.PropTypes.object,
	};

	props: {
		repo: string;
		rev: string;
		path: string;
		overlay: boolean;
		prefetch: ?boolean;
		location: Location;
		route: Route;
		onChangeQuery: (query: string) => void;
		onSelectPath: (path: string) => void;
	};

	state: TreeSearch.props & {
		query: string;
		focused: boolean;
		matchingDefs: {Defs: Array<Def>};
		selectionIndex: number;
	};

	static contextTypes = {
		router: React.PropTypes.object.isRequired,
	};

	constructor(props: TreeSearch.props) {
		super(props);
		this.state = {
			query: "",
			focused: !props.overlay,
			matchingDefs: {Defs: []},
			selectionIndex: 0,
		};
		this._queryInput = null;
		this._handleKeyDown = this._handleKeyDown.bind(this);
		this._scrollToVisibleSelection = this._scrollToVisibleSelection.bind(this);
		this._setSelectedItem = this._setSelectedItem.bind(this);
		this._focusInput = this._focusInput.bind(this);
		this._handleFocus = this._handleFocus.bind(this);
		this._blurInput = this._blurInput.bind(this);
		this._onSelection = debounce(this._onSelection.bind(this), 100, {leading: false, trailing: true}); // Prevent rapid repeated selections
		this._debouncedSetQuery = debounce((query) => {
			if (query !== this.state.query) {
				this.props.onChangeQuery(query);
			}
		}, 75, {leading: false, trailing: true});
	}

	componentDidMount() {
		super.componentDidMount();
		if (global.document) {
			document.addEventListener("keydown", this._handleKeyDown);
		}

		if (global.window) {
			window.addEventListener("focus", this._focusInput);
		}
	}

	componentWillUnmount() {
		super.componentWillUnmount();
		if (global.document) {
			document.removeEventListener("keydown", this._handleKeyDown);
		}
		if (global.window) {
			window.removeEventListener("focus", this._focusInput);
		}
	}

	stores(): Array<Object> { return [TreeStore, DefStore]; }

	reconcileState(state: TreeSearch.state, props: TreeSearch.props): void {
		Object.assign(state, props);

		state.query = props.query || "";

		state.fileTree = TreeStore.fileTree.get(state.repo, state.rev);
		state.fileList = TreeStore.fileLists.get(state.repo, state.rev);

		state.srclibDataVersion = TreeStore.srclibDataVersions.get(state.repo, state.rev);
		state.matchingDefs = state.srclibDataVersion && state.srclibDataVersion.CommitID ? DefStore.defs.list(state.repo, state.srclibDataVersion.CommitID, state.query) : null;
	}

	onStateTransition(prevState: TreeSearch.state, nextState: TreeSearch.state) {
		const prefetch = nextState.prefetch && nextState.prefetch !== prevState.prefetch;
		if (prefetch || nextState.repo !== prevState.repo || nextState.rev !== prevState.rev) {
			Dispatcher.Backends.dispatch(new TreeActions.WantSrclibDataVersion(nextState.repo, nextState.rev));
			Dispatcher.Backends.dispatch(new TreeActions.WantFileList(nextState.repo, nextState.rev));
		}

		if (prevState.srclibDataVersion !== nextState.srclibDataVersion || prevState.query !== nextState.query) {
			if (nextState.srclibDataVersion && nextState.srclibDataVersion.CommitID) {
				Dispatcher.Backends.dispatch(
					new DefActions.WantDefs(nextState.repo, nextState.srclibDataVersion.CommitID, nextState.query)
				);
			}
		}

		if (prevState.fileList !== nextState.fileList) {
			nextState.fuzzyFinder = nextState.fileList ? new Fuze(nextState.fileList.Files, {
				distance: 1000,
				location: 0,
				threshold: 0.1,
			}) : null;
		}

		if (prevState.path !== nextState.path || prevState.query !== nextState.query || prevState.fileList !== nextState.fileList || prevState.fileTree !== nextState.fileTree) {
			nextState.fileResults = null;

			// Show entire file tree as file results.
			if (!nextState.query) {
				if (nextState.fileTree) {
					let dirLevel = nextState.fileTree;
					for (const part of pathSplit(nextState.path)) {
						if (dirLevel.Dirs[part]) {
							dirLevel = dirLevel.Dirs[part];
						} else {
							if (!dirLevel.Dirs[part] && !dirLevel.Files[part]) {
								throw new Error(`invalid path: '${part}'`);
							}
							break;
						}
					}

					const pathPrefix = nextState.path.replace(/^\/$/, "");
					const dirs = Object.keys(dirLevel.Dirs).map(dir => ({
						name: dir,
						isDirectory: true,
						path: `${pathPrefix}/${dir}`,
						url: urlToTree(nextState.repo, nextState.rev, `${pathPrefix}/${dir}`),
					}));
					const files = dirLevel.Files.map(file => ({
						name: file,
						isDirectory: false,
						url: urlToBlob(nextState.repo, nextState.rev, `${pathPrefix}/${file}`),
					}));
					nextState.fileResults = dirs.concat(files);
				}
			} else {
				nextState.selectionIndex = 0;
				if (nextState.fuzzyFinder) {
					nextState.fileResults = nextState.fuzzyFinder.search(nextState.query).map(i => nextState.fileList.Files[i]).map(file => ({
						name: file,
						isDirectory: false,
						url: urlToBlob(nextState.repo, nextState.rev, file),
					}));
				}
			}
		}
	}

	_navigateTo(url: string) {
		this.context.router.push(url);
	}

	_handleKeyDown(e: KeyboardEvent) {
		let idx, max;
		switch (e.keyCode) {
		case 40: // ArrowDown
			idx = this._normalizedSelectionIndex();
			max = this._numResults();

			this.setState({
				selectionIndex: idx + 1 >= max ? 0 : idx + 1,
			}, this._scrollToVisibleSelection);

			e.preventDefault(); // TODO: make keyboard actions scroll automatically with selection
			break;

		case 38: // ArrowUp
			idx = this._normalizedSelectionIndex();
			max = this._numResults();

			this.setState({
				selectionIndex: idx < 1 ? max-1 : idx-1,
			}, this._scrollToVisibleSelection);

			e.preventDefault(); // TODO: make keyboard actions scroll automatically with selection
			break;

		case 37: // ArrowLeft
			if (this.state.path.length !== 0) {
				// Remove last item from path.
				const parts = pathSplit(this.state.path);
				const parentPath = pathJoin(parts.splice(0, parts.length - 1));
				this.state.onSelectPath(parentPath);
			}

			// Allow default (cursor movement in <input>)
			break;

		case 39: // ArrowRight
			this._onSelection();

			// Allow default (cursor movement in <input>)
			break;

		case 13: // Enter
			this._onSelection();
			e.preventDefault();
			break;

		default:
			if (this.state.focused) {
				setTimeout(() => this._debouncedSetQuery(this._queryInput ? this._queryInput.getValue() : ""), 0);
			}
			break;
		}
	}

	_scrollToVisibleSelection() {
		if (this._selectedItem) ReactDOM.findDOMNode(this._selectedItem).scrollIntoView(false);
	}

	_setSelectedItem(e: any) {
		this._selectedItem = e;
	}

	_focusInput() {
		this.setState({focused: true});
		if (this.refs.input) this.refs.input.focus();
	}

	_handleFocus() {
		this._focusInput();
	}

	_blurInput() {
		if (this.refs.input) this.refs.input.blur();
		this.setState({
			focused: false,
		});
	}

	_numSymbolResults(): number {
		if (!this.state.matchingDefs || !this.state.matchingDefs.Defs) return 0;
		return Math.min(this.state.matchingDefs.Defs.length, SYMBOL_LIMIT);
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
		return this._numFileResults() + this._numSymbolResults();
	}

	_normalizedSelectionIndex(): number {
		return Math.min(this.state.selectionIndex, this._numResults() - 1);
	}

	_onSelection() {
		const i = this._normalizedSelectionIndex();
		if (i < this._numSymbolResults()) {
			// Def result
			const def = this.state.matchingDefs.Defs[i];
			this._navigateTo(urlToDef(def));
		} else {
			// File or dir result
			let result = this.state.fileResults[i - this._numSymbolResults()];
			if (result.isDirectory) {
				this.state.onSelectPath(result.path);
			} else {
				this._navigateTo(result.url);
			}
		}
	}

	// returns the selected directory name, or null
	_getSelectedPathPart(): ?string {
		const i = this._normalizedSelectionIndex();
		if (i < this._numSymbolResults()) {
			return null;
		}

		const result = this.state.fileResults[i - this._numSymbolResults()];
		if (result.isDirectory) return result.name;
		return null;
	}

	// returns the selected file name, or null
	_getSelectedFile(): ?string {
		const i = this._normalizedSelectionIndex();
		if (i < this._numSymbolResults()) {
			return null;
		}

		const result = this.state.fileResults[i - this._numSymbolResults()];
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

			const selected = this._normalizedSelectionIndex() - this._numSymbolResults() === i;

			list.push(
				<Link styleName={selected ? "list-item-selected" : "list-item"}
					onMouseOver={() => this._selectItem(i + this._numSymbolResults())}
					ref={selected ? this._setSelectedItem : null}
					to={itemURL}
					key={itemURL}>
					<span style={{paddingRight: "1rem"}}><Icon name={item.isDirectory ? "folder" : "file"} /></span>
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

	_symbolItems(): Array<any> {
		const emptyItem = <div styleName="list-item list-item-empty" key="_nosymbol"><i>No matches.</i></div>;
		if (!this.state.matchingDefs) return [emptyItem];

		if (this.state.matchingDefs && (!this.state.matchingDefs.Defs || this.state.matchingDefs.Defs.length === 0)) return [emptyItem];

		let list = [],
			limit = this.state.matchingDefs.Defs.length > SYMBOL_LIMIT ? SYMBOL_LIMIT : this.state.matchingDefs.Defs.length;

		for (let i = 0; i < limit; i++) {
			let def = this.state.matchingDefs.Defs[i];
			let defURL = urlToDef(def);

			const selected = this._normalizedSelectionIndex() === i;

			list.push(
				<Link styleName={selected ? "list-item-selected" : "list-item"}
					onMouseOver={() => this._selectItem(i)}
					ref={selected ? this._setSelectedItem : null}
					to={defURL}
					key={defURL}>
					<code>{qualifiedNameAndType(def)}</code>
				</Link>
			);
		}

		return list;
	}

	_overlayBreadcrumb() {
		const urlToPathPrefix = (i) => {
			const parts = pathSplit(this.state.path);
			const pathPrefix = pathJoin(parts.splice(0, i + 1));
			return urlToTree(this.state.repo, this.state.rev, pathPrefix);
		};

		let filepath = this.state.path;
		if (filepath.indexOf("/") === 0) filepath = filepath.substring(1);

		let fileBreadcrumb = breadcrumb(
			filepath,
			(i) => <span key={i} styleName="path-sep">/</span>,
			(path, component, i, isLast) => (
				<Link to={urlToPathPrefix(i)}
					key={i}
					styleName={isLast ? "path-active" : "path-inactive"}>
					{component}
				</Link>
			),
		);

		return (
			<span styleName="file-path">
				<RepoLink repo={`${this.state.repo}/`} />
				{fileBreadcrumb}
			</span>
		);
	}

	render() {
		return (
			<div styleName="tree-common">
				<div styleName="input-container">
					<Input type="text"
						block={true}
						onFocus={this._focusInput}
						onBlur={this._blurInput}
						autoFocus={true}
						defaultValue={this.state.query}
						placeholder="Jump to symbols or files..."
						ref={(c) => this._queryInput = c} />
				</div>
				<div styleName="list-header">
					Symbols
				</div>
				<div>
					{this.state.srclibDataVersion && this.state.srclibDataVersion.CommitID && this._symbolItems()}
					{this.state.srclibDataVersion && !this.state.srclibDataVersion.CommitID &&
						<div styleName="list-item list-item-empty">
							<span style={{paddingRight: "1rem"}}><Loader /></span>
							<i>Sourcegraph is analyzing your code &mdash;
								<Link styleName="link" to={urlToBuilds(this.state.repo)}>results will be available soon!</Link>
							</i>
						</div>
					}
				</div>
				<div styleName="list-header">
					Files
					{!this.state.query && this.state.overlay && this._overlayBreadcrumb()}
				</div>
				<div styleName="list-item-group">
					{this._listItems()}
				</div>
			</div>
		);
	}
}

export default CSSModules(TreeSearch, styles, {allowMultiple: true});
