// @flow

import React from "react";
import ReactDOM from "react-dom";
import {Link} from "react-router";
import fuzzysearch from "fuzzysearch";
import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import debounce from "lodash/function/debounce";
import trimLeft from "lodash/string/trimLeft";
import TreeStore from "sourcegraph/tree/TreeStore";
import DefStore from "sourcegraph/def/DefStore";
import SearchStore from "sourcegraph/search/SearchStore";
import "sourcegraph/tree/TreeBackend";
import "sourcegraph/def/DefBackend";
import "sourcegraph/search/SearchBackend";
import * as TreeActions from "sourcegraph/tree/TreeActions";
import * as DefActions from "sourcegraph/def/DefActions";
import * as SearchActions from "sourcegraph/search/SearchActions";
import Header from "sourcegraph/components/Header";
import {qualifiedNameAndType} from "sourcegraph/def/Formatter";
import {urlToBlob} from "sourcegraph/blob/routes";
import {urlToDef} from "sourcegraph/def/routes";
import {urlToTree} from "sourcegraph/tree/routes";
import httpStatusCode from "sourcegraph/util/httpStatusCode";
import {urlToBuilds} from "sourcegraph/build/routes";
import type {Def} from "sourcegraph/def";
import type {Route} from "react-router";

import {Input, Loader, RepoLink} from "sourcegraph/components";
import {FileIcon, FolderIcon} from "sourcegraph/components/Icons";

import breadcrumb from "sourcegraph/util/breadcrumb";

import CSSModules from "react-css-modules";
import styles from "./styles/Tree.css";

const SYMBOL_LIMIT = 5;
const GLOBAL_DEFS_LIMIT = 3;
const LOCAL_DEFS_LIMIT = 10;
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
	matchingDefs: ?{Defs: Array<Def>};
	lastDefinedMatchingDefs?: ?{Defs: Array<Def>};
	xdefs: ?{Defs: Array<Def>};
	defListFilePathPrefix: ?string;
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
			matchingDefs: null,
			xdefs: null,
			selectionIndex: 0,
			defListFilePathPrefix: null,
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
		if (this._srclibBuildingInterval) clearInterval(this._srclibBuildingInterval);
	}

	stores(): Array<Object> { return [TreeStore, DefStore, SearchStore]; }

	reconcileState(state: TreeSearchState, props: TreeSearchProps): void {
		// $FlowHack
		Object.assign(state, props);

		state.query = props.query || "";

		state.fileTree = TreeStore.fileTree.get(state.repo, state.commitID);
		state.fileList = TreeStore.fileLists.get(state.repo, state.commitID);

		// Limit defs to the current directory unless we're querying. That
		// should be global to be consistent with file list behavior (for which
		// searches are global).
		state.defListFilePathPrefix = state.query || state.path === "/" ? null : `${state.path}/`;
		state.srclibDataVersion = TreeStore.srclibDataVersions.get(state.repo, state.commitID);

		// $FlowHack: this.context.user.Admin
		if (!this.context.user.Admin) { // DEPRECATED: remove after search index successful
			state.matchingDefs = state.srclibDataVersion && state.srclibDataVersion.CommitID ? DefStore.defs.list(state.repo, state.srclibDataVersion.CommitID, state.query, state.defListFilePathPrefix) : null;
		} else if (state.rev && state.rev.length === 40) {
			state.matchingDefs = SearchStore.get(state.query, [this.props.repo], null, state.rev, LOCAL_DEFS_LIMIT);
		} else {
			state.matchingDefs = SearchStore.get(state.query, [this.props.repo], null, null, LOCAL_DEFS_LIMIT);
		}
		state.xdefs = SearchStore.get(state.query, null, [this.props.repo], null, GLOBAL_DEFS_LIMIT);
	}

	onStateTransition(prevState: TreeSearchState, nextState: TreeSearchState) {
		const prefetch = nextState.prefetch && nextState.prefetch !== prevState.prefetch;
		if (prefetch || nextState.repo !== prevState.repo || nextState.commitID !== prevState.commitID) {
			if (nextState.commitID) {
				Dispatcher.Backends.dispatch(new TreeActions.WantSrclibDataVersion(nextState.repo, nextState.commitID));
				Dispatcher.Backends.dispatch(new TreeActions.WantFileList(nextState.repo, nextState.commitID));
			}
		}

		// If there was previously a response from the server but no srclib
		// data version, i.e., if the repository has not been built recently
		// then poll against the server for an update periodically.
		const pollSrclibVersion = nextState.srclibDataVersion && !nextState.srclibDataVersion.CommitID;
		if (pollSrclibVersion && nextState.commitID && !this._srclibBuildingInterval) {
			const pollInterval = 500;
			this._srclibBuildingInterval = setInterval(() => {
				Dispatcher.Backends.dispatch(new TreeActions.WantSrclibDataVersion(nextState.repo, nextState.commitID, null, true));
			}, pollInterval);
		} else if (!pollSrclibVersion && this._srclibBuildingInterval) {
			clearInterval(this._srclibBuildingInterval);
			this._srclibBuildingInterval = null;
		}

		// $FlowHack: this.context.user.Admin
		if (!this.context.user.Admin) { // DEPRECATED: remove after search index successful
			if (prevState.srclibDataVersion !== nextState.srclibDataVersion || prevState.query !== nextState.query || prevState.defListFilePathPrefix !== nextState.defListFilePathPrefix) {
				// Only fetch on the client, not server, so that we don't
				// cache stale def lists prior to the repo's first build.
				if (typeof document !== "undefined" && nextState.srclibDataVersion && nextState.srclibDataVersion.CommitID) {
					Dispatcher.Backends.dispatch(
						new DefActions.WantDefs(nextState.repo, nextState.srclibDataVersion.CommitID, nextState.query, nextState.defListFilePathPrefix, nextState.overlay || false),
					);
				}
			}
		} else {
			if (prevState.query !== nextState.query || prevState.repo !== nextState.repo) {
				if (nextState.rev && nextState.rev.length === 40) {
					Dispatcher.Backends.dispatch(new SearchActions.WantResults({
						query: nextState.query,
						repos: [nextState.repo],
						commitID: nextState.rev,
						limit: LOCAL_DEFS_LIMIT,
					}));
				} else {
					Dispatcher.Backends.dispatch(new SearchActions.WantResults({
						query: nextState.query,
						repos: [nextState.repo],
						limit: LOCAL_DEFS_LIMIT,
					}));
				}
			}
			Dispatcher.Backends.dispatch(new SearchActions.WantResults({
				query: nextState.query,
				notRepos: [nextState.repo],
				limit: GLOBAL_DEFS_LIMIT,
			}));
		}

		if (prevState.matchingDefs && prevState.matchingDefs !== nextState.matchingDefs) {
			nextState.lastDefinedMatchingDefs = prevState.matchingDefs;
		}

		if (prevState.matchingDefs !== nextState.matchingDefs) {
			// Keep selectionIndex on same file item even after def results are loaded. Prevents
			// the selection from jumping around as more data comes in.
			const prevNumDefs = Math.min(SYMBOL_LIMIT, prevState.matchingDefs && prevState.matchingDefs.Defs ? prevState.matchingDefs.Defs.filter(this._symbolFilter).length : 0);
			const nextNumDefs = Math.min(SYMBOL_LIMIT, nextState.matchingDefs && nextState.matchingDefs.Defs ? nextState.matchingDefs.Defs.filter(this._symbolFilter).length : 0);
			const defWasSelected = nextState.selectionIndex < prevNumDefs || (prevState.fileResults && prevState.fileResults.length === 0);
			if (defWasSelected) {
				nextState.selectionIndex = 0;
			} else {
				nextState.selectionIndex += (nextNumDefs - prevNumDefs);
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
					// $FlowHack: nextState.path is non-null
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

					// $FlowHack: nextState.path is non-null
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

			} else if (nextState.fileList && nextState.fileList.Files) {
				nextState.fileResults = nextState.fileList.Files
					.filter((f) => fuzzysearch(nextState.query, f))
					.map((f) => ({
						name: f,
						isDirectory: false,
						url: urlToBlob(nextState.repo, nextState.rev, f),
					}));
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

	_numSymbolResults(): number {
		if (!this.state.matchingDefs || !this.state.matchingDefs.Defs) return 0;
		return Math.min(this.state.matchingDefs.Defs.filter(this._symbolFilter).length, SYMBOL_LIMIT);
	}

	_numXDefResults(): number {
		if (!this.state.xdefs || !this.state.xdefs.Defs) return 0;
		return this.state.xdefs.Defs.length;
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
		return this._numFileResults() + this._numSymbolResults() + this._numXDefResults();
	}

	_normalizedSelectionIndex(): number {
		return Math.min(this.state.selectionIndex, this._numResults() - 1);
	}

	_onSelection() {
		const i = this._normalizedSelectionIndex();
		if (i < this._numSymbolResults()) {
			// Def result
			if (!this.state.matchingDefs || !this.state.matchingDefs.Defs || this.state.matchingDefs.Defs.length === 0) return;
			const def = this.state.matchingDefs.Defs.filter(this._symbolFilter)[i];
			this._navigateTo(urlToDef(def, this.state.rev));
		} else if (i >= this._numSymbolResults() && i < this._numSymbolResults() + this._numXDefResults()) {
			// XDef result
			if (!this.state.xdefs || !this.state.xdefs.Defs || this.state.xdefs.Defs.length === 0) return;
			let d = i - this._numSymbolResults();
			// $FlowHack: this.state.xdefs is non-null
			const def = this.state.xdefs.Defs[d];
			this._navigateTo(urlToDef(def));
		} else {
			// File or dir result
			if (!this.state.fileResults) return;
			let d = i - (this._numSymbolResults() + this._numXDefResults());
			let result = this.state.fileResults[d];
			if (!result) return;
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

	_listItems(offset: number): Array<any> {
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

			const selected = this._normalizedSelectionIndex() - offset === i;

			let icon;
			if (item.isParentDirectory) icon = null;
			else if (item.isDirectory) icon = <FolderIcon styleName="icon" />;
			else icon = <FileIcon styleName="icon" />;

			let key = `f:${itemURL}`;
			list.push(
				<Link styleName={`${selected ? "list-item-selected" : "list-item"} ${item.isParentDirectory ? "parent-dir" : ""}`}
					onMouseOver={(ev) => this._mouseSelectItem(ev, i + this._numSymbolResults() + this._numXDefResults())}
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

	_symbolFilter(def) {
		// Do not show package results.
		return def.Kind !== "package";
	}

	_symbolItems(offset: number): ?Array<any> {
		if (!this.state.srclibDataVersion || !this.state.srclibDataVersion.CommitID) return null;

		const contentPlaceholderItem = (i) => (
			<div styleName="list-item" key={`_nosymbol${i}`}>
				<div styleName="content-placeholder" style={{width: `${20 + ((i+1)*39)%60}%`}}><code>&nbsp;</code></div>
			</div>
		);
		if (!this.state.matchingDefs) {
			let numPlaceholders = 1;
			if (!this.state.query) {
				numPlaceholders = SYMBOL_LIMIT;
			} else if (this.state.lastDefinedMatchingDefs && this.state.lastDefinedMatchingDefs.Defs) {
				numPlaceholders = Math.min(Math.max(1, this.state.lastDefinedMatchingDefs.Defs.length), SYMBOL_LIMIT);
			}
			const placeholders = [];
			for (let i = 0; i < numPlaceholders; i++) {
				placeholders.push(contentPlaceholderItem(i));
			}
			return placeholders;
		}

		if (this.state.matchingDefs && (!this.state.matchingDefs.Defs || this.state.matchingDefs.Defs.filter(this._symbolFilter).length === 0)) {
			return [<div styleName="list-item list-item-empty" key="_nosymbol"><i>No matches.</i></div>];
		}

		const defs = this.state.matchingDefs.Defs.filter(this._symbolFilter);
		let list = [],
			limit = defs.length > SYMBOL_LIMIT ? SYMBOL_LIMIT : defs.length;
		for (let i = 0; i < limit; i++) {
			let def = defs[i];
			list.push(this._defToLink(def, this.state.rev, offset + i, "i"));
		}

		return list;
	}

	_xdefItems(offset: number): {items: ?Array<any>, count: number} {
		let groups = [];
		let groupToDefs = {};

		if (!this.state.xdefs || !this.state.xdefs.Defs) {
			return {items: [], count: 0};
		}
		let defs = this.state.xdefs.Defs;
		for (let i = 0; i < defs.length; i++) {
			let repo = defs[i].Repo;
			if (!groupToDefs[repo]) {
				let repoDefs = [];
				groups.push(repo);
				groupToDefs[repo] = repoDefs;
			}
			groupToDefs[repo].push(defs[i]);
		}

		let sections = [];
		let idx = offset;
		for (let i = 0; i < groups.length; i++) {
			let items = [];
			let repo = groups[i];
			let rdefs = groupToDefs[repo];
			for (let j = 0; j < rdefs.length; j++) {
				let def = rdefs[j];
				items.push(this._defToLink(def, null, idx, "x"));
				idx++;
			}
			sections.push(
				<div key={`group-header:${repo}`} styleName="list-header">Symbols in {repo}</div>,
				<div key={`group:${repo}`} styleName="list-item-group">
					{items}
				</div>
			);
		}
		return {items: sections, count: idx - offset};
	}

	_defToLink(def: Def, rev: ?string, i: number, prefix: string) {
		const selected = this._normalizedSelectionIndex() === i;
		let defURL = urlToDef(def, rev);
		let key = `${prefix}:${defURL}`;
		return (
				<Link styleName={selected ? "list-item-selected" : "list-item"}
					onMouseOver={(ev) => this._mouseSelectItem(ev, i)}
					ref={selected ? this._setSelectedItem : null}
					to={defURL}
					key={key}>
						<code>{qualifiedNameAndType(def)}</code>
				</Link>
		);
	}

	_overlayBreadcrumb() {
		if (!this.state.path) return null;
		let path = this.state.path;

		const urlToPathPrefix = (i) => {
			const parts = pathSplit(path);
			const pathPrefix = pathJoin(parts.splice(0, i + 1));
			return urlToTree(this.state.repo, this.state.rev, pathPrefix);
		};

		let filepath = path;
		if (filepath.indexOf("/") === 0) filepath = filepath.substring(1);

		let fileBreadcrumb = breadcrumb(
			filepath,
			(i) => <span key={i} styleName="path-sep">/</span>,
			(p, component, i, isLast) => (
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
		if (this.state.fileResults && this.state.fileResults.Error) {
			let code = httpStatusCode(this.state.fileResults.Error);
			return (
				<Header
					title={`${code}`}
					subtitle={code === 404 ? `Directory not found.` : "Directory is not available."} />
			);
		}

		let symbolItems = this._symbolItems(0) || [];
		let xdefInfo = this._xdefItems(this._numSymbolResults()) || {items: [], count: 0};
		let listItems = this._listItems(this._numSymbolResults() + this._numXDefResults()) || [];

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
						placeholder="Jump to symbols or files..."
						spellCheck={false}
						domRef={(e) => this._queryInput = e} />
				</div>
				<div styleName="list-header">
					Symbols in current repository
				</div>
				<div>
					{symbolItems}
					{this.state.srclibDataVersion && !this.state.srclibDataVersion.CommitID &&
						<div styleName="list-item list-item-empty">
							<span style={{paddingRight: "1rem"}}><Loader /></span>
							<i>Sourcegraph is analyzing your code &mdash;
								<Link styleName="link" to={urlToBuilds(this.state.repo)}>results will be available soon!</Link>
							</i>
						</div>
					}
				</div>

				{xdefInfo.items}

				<div styleName="list-header">
					Files in
					{!this.state.query && this.state.overlay && this._overlayBreadcrumb()}
				</div>
				<div styleName="list-item-group">
					{listItems}
				</div>
			</div>
		);
	}
}

export default CSSModules(TreeSearch, styles, {allowMultiple: true});
