import React from "react";
import URL from "url";
import Fuze from "fuse.js";
import classNames from "classnames";
import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import debounce from "lodash/function/debounce";
import * as router from "sourcegraph/util/router";
import TreeStore from "sourcegraph/tree/TreeStore";
import SearchResultsStore from "sourcegraph/search/SearchResultsStore";
import "sourcegraph/tree/TreeBackend";
import * as TreeActions from "sourcegraph/tree/TreeActions";
import * as SearchActions from "sourcegraph/search/SearchActions";

import TreeStyles from "./styles/Tree.css";
import BaseStyles from "sourcegraph/components/styles/base.css";

const SYMBOL_LIMIT = 5;
const FILE_LIMIT = 5;

class TreeSearch extends Container {
	constructor(props) {
		super(props);
		this.state = {
			visible: false,
			loading: false,
			matchingSymbols: {Results: [], SrclibDataVersion: null},
			matchingFiles: [],
			fileResults: [], // either the full directory tree (for empty query) or matching file paths
			query: "",
			selectionIndex: 0,
		};
		this._handleKeyUp = this._handleKeyUp.bind(this);
		this._focusInput = this._focusInput.bind(this);
		this._blurInput = this._blurInput.bind(this);
		this._onType = this._onType.bind(this);
		this._currentNumResults = this._currentNumResults.bind(this);
		this._getSelectionURL = this._getSelectionURL.bind(this);
		this._debouncedSetQuery = debounce((query) => {
			this.setState({query: query, selectionIndex: 0});
		}, 75, {leading: false, trailing: true});
	}

	componentDidMount() {
		super.componentDidMount();
		if (!this.state.overlay) {
			this._focusInput();
		} else {
			document.addEventListener("keyup", this._handleKeyUp);
		}
	}

	componentWillUnmount() {
		super.componentWillUnmount();
		if (this.state.overlay) {
			document.removeEventListener("keyup", this._handleKeyUp);
		}
	}

	stores() { return [TreeStore, SearchResultsStore]; }

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	onStateTransition(prevState, nextState) {
		let fileList = TreeStore.fileLists.get(nextState.repo, nextState.rev);
		nextState.allFiles = fileList ? fileList.Files : null;
		nextState.fileTree = TreeStore.fileTree.get(nextState.repo, nextState.rev);
		nextState.matchingSymbols =
			SearchResultsStore.results.get(nextState.repo, nextState.rev, nextState.query, "tokens", 1) ||
			{Results: [], SrclibDataVersion: null};

		const becameVisible = nextState.visible && nextState.visible !== prevState.visible;
		if (becameVisible || nextState.repo !== prevState.repo || nextState.rev !== prevState.rev) {
			// Don't load the file list when the page loads until we become visible.
			const initialLoad = !prevState.repo && !prevState.rev;
			if (!initialLoad || nextState.prefetch) {
				Dispatcher.asyncDispatch(new TreeActions.WantFileList(nextState.repo, nextState.rev));
				Dispatcher.asyncDispatch(
					new SearchActions.WantResults(nextState.repo, nextState.rev, "tokens", 1, SYMBOL_LIMIT, nextState.query)
				);
			}
		}

		if (nextState.allFiles !== prevState.allFiles) {
			nextState.fuzzyFinder = nextState.allFiles && new Fuze(nextState.allFiles, {
				distance: 1000,
				location: 0,
				threshold: 0.1,
			});
			nextState.loading = nextState.allFiles === null;
		}

		if (nextState.fuzzyFinder !== prevState.fuzzyFinder || nextState.query !== prevState.query) {
			nextState.matchingFiles = (nextState.query && nextState.fuzzyFinder) ?
			nextState.fuzzyFinder.search(nextState.query).map(i => nextState.allFiles[i]) :
			nextState.allFiles;
		}

		if (nextState.query !== prevState.query) {
			Dispatcher.asyncDispatch(
				new SearchActions.WantResults(nextState.repo, nextState.rev, "tokens", 1, SYMBOL_LIMIT, nextState.query)
			);
		}

		if (nextState.query === "") {
			// Show entire file tree as file results.
			if (nextState.fileTree) {
				let dirLevel = nextState.fileTree;
				for (const part of nextState.currPath) {
					if (dirLevel.Dirs[part]) {
						dirLevel = dirLevel.Dirs[part];
					} else {
						break;
					}
				}
				const dirs = Object.keys(dirLevel.Dirs).map(dir => ({
					name: dir,
					isDirectory: true,
					url: router.tree(this.props.repo, this.props.rev,
						`${nextState.currPath.join("/")}${nextState.currPath.length > 0 ? "/" : ""}${dir}`),
				}));

				nextState.fileResults = dirs.concat(dirLevel.Files.map(file => ({
					name: file,
					isDirectory: false,
					url: router.tree(this.props.repo, this.props.rev,
						`${nextState.currPath.join("/")}${nextState.currPath.length > 0 ? "/" : ""}${file}`),
				})));
			}
		} else if (nextState.matchingFiles && nextState.matchingFiles !== prevState.matchingFiles) {
			nextState.fileResults = nextState.matchingFiles.map(file => ({
				name: file,
				isDirectory: false,
				url: router.tree(this.props.repo, this.props.rev, file),
			}));
		}

		// Prevent out-of-bounds, e.g. after directory navigation.
		if (nextState.matchingSymbols !== prevState.matchingSymbols ||
			nextState.fileResults !== prevState.fileResults) {
			const nextNumResults =
				this._numResults(nextState.query, nextState.fileResults || [], nextState.matchingSymbols || []);
			const prevNumResults =
				this._numResults(prevState.query, prevState.fileResults || [], prevState.matchingSymbols || []);
			if (nextNumResults < prevNumResults) {
				if (nextState.selectionIndex >= nextNumResults) {
					nextState.selectionIndex = nextNumResults - 1;
				}
			}
		}
	}

	_handleKeyUp(ev) {
		const tag = ev.target.tagName;
		switch (ev.keyCode) {
		case 84: // "t"
			if (tag === "INPUT" || tag === "SELECT" || tag === "TEXTAREA") return;
			this._focusInput();
			break;

		case 27: // ESC
			this._blurInput();
		}
	}

	_focusInput() {
		if (document.body.dataset.fileSearchDisabled) {
			return null;
		}

		this.setState({
			visible: true,
			selectionIndex: 0,
		}, () => this.refs.input && this.refs.input.focus());
	}

	_blurInput() {
		if (this.refs.input) this.refs.input.blur();

		this.setState({
			visible: false,
			loading: false,
		});
	}

	_numResults(query, fileResults, matchingSymbols) {
		let numFileResults = fileResults.length > FILE_LIMIT ? FILE_LIMIT : fileResults.length;
		if (query === "") numFileResults = fileResults.length; // override to show full directory tree on empty query
		const numSymbolResults = matchingSymbols.Results.length > SYMBOL_LIMIT ? SYMBOL_LIMIT : matchingSymbols.Results.length;
		return numFileResults + numSymbolResults;
	}

	_currentNumResults() {
		return this._numResults(this.state.query, this.state.fileResults, this.state.matchingSymbols);
	}

	_getSelectionURL() {
		const i = this.state.selectionIndex;
		if (i < this.state.matchingSymbols.Results.length) {
			const def = this.state.matchingSymbols.Results[i].Def;
			return router.def(def.Repo, def.CommitID, def.UnitType, def.Unit, def.Path);
		}
		return this.state.fileResults[i - this.state.matchingSymbols.Results.length].url;
	}

	// returns the selected directory name, or null
	_getSelectedPathPart() {
		const i = this.state.selectionIndex;
		if (i < this.state.matchingSymbols.Results.length) {
			return null;
		}

		const result = this.state.fileResults[i - this.state.matchingSymbols.Results.length];
		if (result.isDirectory) return result.name;
		return null;
	}

	_onType(e) {
		let idx, max, part;
		switch (e.key) {
		case "ArrowDown":
			idx = this.state.selectionIndex;
			max = this._currentNumResults();

			this.setState({
				selectionIndex: idx + 1 >= max ? 0 : idx + 1,
			});

			e.preventDefault();
			break;

		case "ArrowUp":
			idx = this.state.selectionIndex;
			max = this._currentNumResults();

			this.setState({
				selectionIndex: idx < 1 ? max-1 : idx-1,
			});

			e.preventDefault();
			break;

		case "ArrowLeft":
			if (this.state.currPath.length !== 0) {
				Dispatcher.dispatch(new TreeActions.UpDirectory());
			}

			e.preventDefault();
			break;

		case "ArrowRight":
			part = this._getSelectedPathPart();
			if (part) {
				Dispatcher.dispatch(new TreeActions.DownDirectory(part));
			}

			e.preventDefault();
			break;

		case "Enter":
			window.location = this._getSelectionURL();
			e.preventDefault();
			break;

		default:
			this._debouncedSetQuery(this.refs.input ? this.refs.input.value : "");
		}
	}

	_listItems() {
		const items = this.state.fileResults;
		if (!this.state.visible || !items || items.length === 0) return [<div className={TreeStyles.list_item} key="_nofiles"><i>No matches!</i></div>];

		let list = [],
			limit = items.length > FILE_LIMIT ? FILE_LIMIT : items.length;

		// Override limit if query is empty to show the full directory tree.
		if (this.state.query === "") limit = items.length;

		for (let i = 0; i < limit; i++) {
			let item = items[i],
				itemURL = item.url;

			const selected = this.state.selectionIndex - this.state.matchingSymbols.Results.length === i;

			list.push(
				<div className={selected ? TreeStyles.list_item_selected : TreeStyles.list_item} key={itemURL}>
					<span className={TreeStyles.filetype_icon}><i className={classNames("fa", {
						"fa-file-text-o": !item.isDirectory,
						"fa-folder": item.isDirectory,
					})}></i>
					</span>
					<a className={BaseStyles.link} href={itemURL}>{item.name}</a>
					{this.state.query === "" && item.isDirectory &&
						<span className={TreeStyles.directory_nav_icon}>
						<i className="fa fa-chevron-right" />
						</span>
					}
				</div>
			);
		}

		return list;
	}

	_symbolItems() {
		if (!this.state.visible || !this.state.matchingSymbols || this.state.matchingSymbols.Results.length === 0) {
			return [<div className={TreeStyles.list_item} key="_nosymbol"><i>No matches!</i></div>];
		}

		let list = [],
			limit = this.state.matchingSymbols.Results.length > SYMBOL_LIMIT ? SYMBOL_LIMIT : this.state.matchingSymbols.Results.length;

		for (let i = 0; i < limit; i++) {
			let result = this.state.matchingSymbols.Results[i];
			let def = result.Def,
				defURL = router.def(def.Repo, def.CommitID, def.UnitType, def.Unit, def.Path);

			const selected = this.state.selectionIndex === i;

			list.push(
				<div className={selected ? TreeStyles.list_item_selected : TreeStyles.list_item} key={defURL}>
					<div key={defURL}>
						<a className={BaseStyles.link} href={defURL}>
							<code>{def.Kind}</code>
							<code dangerouslySetInnerHTML={result.QualifiedName}></code>
						</a>
					</div>
				</div>
			);
		}

		return list;
	}

	_buildsURL() {
		if (window && window.location && window.location.href) {
			let url = URL.parse(window.location.href);
			url.pathname = `${this.state.repo}/.builds`;
			return URL.format(url);
		}

		return "";
	}

	render() {
		return (
			<div className={this.state.visible ? TreeStyles.tree : BaseStyles.hidden}>
				<div className={this.state.overlay ? BaseStyles.overlay : BaseStyles.hidden} onClick={this._blurInput} />
				<div className={this.state.overlay ? TreeStyles.tree_modal : TreeStyles.tree}>
					<div className={TreeStyles.input_group}>
						<input className={TreeStyles.input}
							type="text"
							placeholder="Search this repository..."
							ref="input"
							onKeyUp={this._onType} />
						{!this.state.overlay &&
							<div className={TreeStyles.search_hotkey}>t</div>}
					</div>
					<div className={TreeStyles.list_header}>
						Symbols
					</div>
					<div>
						{this.state.matchingSymbols.SrclibDataVersion && this._symbolItems()}
						{!this.state.matchingSymbols.SrclibDataVersion &&
							<div className={TreeStyles.list_item}>
								<span className={TreeStyles.icon}><i className="fa fa-spinner fa-spin"></i></span>
								<i>Sourcegraph is analyzing your code &mdash;&nbsp;
									<a className={BaseStyles.link} href={this._buildsURL()}>results will be available soon!</a>
								</i>
							</div>
						}
					</div>
					<div className={TreeStyles.list_header}>
						Files
						{this.state.query === "" &&
							<span className={TreeStyles.file_path}>{this.state.currPath.join("/")}</span>}
					</div>
					<div>
						{this._listItems()}
					</div>
				</div>
			</div>
		);
	}
}

TreeSearch.propTypes = {
	repo: React.PropTypes.string.isRequired,
	rev: React.PropTypes.string.isRequired,
	currPath: React.PropTypes.arrayOf(React.PropTypes.string).isRequired,
	overlay: React.PropTypes.bool.isRequired,
	prefetch: React.PropTypes.bool,
};

export default TreeSearch;
