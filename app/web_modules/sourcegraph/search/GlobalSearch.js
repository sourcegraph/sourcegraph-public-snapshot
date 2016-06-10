// @flow

import React from "react";
import ReactDOM from "react-dom";
import {Link} from "react-router";
import {browserHistory} from "react-router";
import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import SearchStore from "sourcegraph/search/SearchStore";
import "sourcegraph/search/SearchBackend";
import debounce from "lodash/function/debounce";
import trimLeft from "lodash/string/trimLeft";
import * as SearchActions from "sourcegraph/search/SearchActions";
import {qualifiedNameAndType} from "sourcegraph/def/Formatter";
import {urlToDef, urlToDefInfo} from "sourcegraph/def/routes";
import type {Repo, Def} from "sourcegraph/def";

import {Input, Icon} from "sourcegraph/components";

import CSSModules from "react-css-modules";
import styles from "./styles/GlobalSearch.css";
import base from "sourcegraph/components/styles/_base.css";

export const RESULTS_LIMIT = 20;

// GlobalSearch is the global search bar + results component.
// Tech debt: this duplicates a lot of code with TreeSearch and we
// should consider merging them at some point.
class GlobalSearch extends Container {
	static propTypes = {
		location: React.PropTypes.object.isRequired,
		query: React.PropTypes.string.isRequired,
	};

	static contextTypes = {
		router: React.PropTypes.object.isRequired,
		eventLogger: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);

		this.state = {
			query: "",
			matchingResults: {Repos: [], Defs: []},
			selectionIndex: 0,
			focused: false,
		};
		this._queryInput = null;
		this._handleKeyDown = this._handleKeyDown.bind(this);
		this._scrollToVisibleSelection = this._scrollToVisibleSelection.bind(this);
		this._setSelectedItem = this._setSelectedItem.bind(this);
		this._handleInput = this._handleInput.bind(this);
		this._onSelection = debounce(this._onSelection.bind(this), 100, {leading: false, trailing: true}); // Prevent rapid repeated selections
		this._onChangeQuery = this._onChangeQuery.bind(this);
		this._debouncedSetQuery = debounce((query) => {
			if (query !== this.state.query) {
				this._onChangeQuery(query);
			}
		}, 200, {leading: false, trailing: true});
	}

	state: {
		query: string;
		matchingResults: {
			Repos: Array<Repo>,
			Defs: Array<Def>
		};
		focused: boolean;
		selectionIndex: number;
	};

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

	stores(): Array<Object> { return [SearchStore]; }

	reconcileState(state: GlobalSearch.state, props) {
		Object.assign(state, props);
		state.matchingResults = SearchStore.results.get(state.query, null, null, RESULTS_LIMIT,
			this.props.location.query.prefixMatch, this.props.location.query.includeRepos);
	}

	onStateTransition(prevState, nextState) {
		if (nextState.query && prevState.query !== nextState.query) {
			Dispatcher.Backends.dispatch(new SearchActions.WantResults(nextState.query, null, null, RESULTS_LIMIT,
			this.props.location.query.prefixMatch, this.props.location.query.includeRepos));
		}
	}

	_onChangeQuery(query: string) {
		this.context.router.replace({...this.props.location, query: {
			q: query || undefined, // eslint-disable-line no-undefined
			prefixMatch: this.props.location.query.prefixMatch || undefined, // eslint-disable-line no-undefined
			includeRepos: this.props.location.query.includeRepos || undefined}}); // eslint-disable-line no-undefined
		this.setState({query: query});
		this.context.eventLogger.logEvent("GlobalSearchInitiated", {globalSearchQuery: query});
	}

	_navigateTo(url: string) {
		browserHistory.push(url);
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

		case 37: // ArrowLeft
			this._temporarilyIgnoreMouseSelection();

			// Allow default (cursor movement in <input>)
			break;

		case 39: // ArrowRight
			this._temporarilyIgnoreMouseSelection();

			// Allow default (cursor movement in <input>)
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

	_numResults(): number {
		if (!this.state.matchingResults ||
			(!this.state.matchingResults.Defs && !this.state.matchingResults.Repos)) return 0;

		let count = 0;
		if (this.state.matchingResults.Defs) {
			count = Math.min(this.state.matchingResults.Defs.length, RESULTS_LIMIT);
		}

		if (this.state.matchingResults.Repos) {
			count += this.state.matchingResults.Repos.length;
		}
		return count;
	}

	_normalizedSelectionIndex(): number {
		return Math.min(this.state.selectionIndex, this._numResults() - 1);
	}

	_onSelection() {
		const i = this._normalizedSelectionIndex();
		if (i === -1) {
			return;
		}

		let offset = 0;
		if (this.state.matchingResults.Repos) {
			if (i < this.state.matchingResults.Repos.length) {
				const url = `/${this.state.matchingResults.Repos[i].URI}`;
				this.context.eventLogger.logEvent("GlobalSearchItemSelected", {globalSearchQuery: this.state.query, selectedItem: url});
				this._navigateTo(url);
				return;
			}

			offset = this.state.matchingResults.Repos.length;
		}

		const def = this.state.matchingResults.Defs[i - offset];
		const url = urlToDefInfo(def) ? urlToDefInfo(def) : urlToDef(def);
		this.context.eventLogger.logEvent("GlobalSearchItemSelected", {globalSearchQuery: this.state.query, selectedItem: url});
		this._navigateTo(url);
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

	_results(): Array<any> {
		if (!this.state.query) return [<div styleName="result" key="_nosymbol"></div>];

		const noResultsItem = <div styleName="tc f4" className={base.pv5} key="_nosymbol">Sorry, we couldn't find anything.</div>;
		if (!this.state.matchingResults) {
			return [<div key="1" styleName="tc f4" className={base.pv5}>Loading results...</div>];
		}

		if (this.state.matchingResults &&
			(!this.state.matchingResults.Defs || this.state.matchingResults.Defs.length === 0) &&
			(!this.state.matchingResults.Repos || this.state.matchingResults.Repos.length === 0)) return [noResultsItem];

		let list = [], numDefs = 0,
			numRepos = this.state.matchingResults.Repos ? this.state.matchingResults.Repos.length : 0;

		if (this.state.matchingResults.Defs) {
			numDefs = this.state.matchingResults.Defs.length > RESULTS_LIMIT ? RESULTS_LIMIT : this.state.matchingResults.Defs.length;
		}

		for (let i = 0; i < numRepos; i++) {
			let repo = this.state.matchingResults.Repos[i];
			const selected = this._normalizedSelectionIndex() === i;

			const firstLineDocString = firstLine(repo.Description);
			list.push(
				<Link styleName={selected ? "block result-selected" : "block result"}
					onMouseOver={(ev) => this._mouseSelectItem(ev, i)}
					ref={selected ? this._setSelectedItem : null}
					to={repo.URI}
					key={repo.URI}
					onClick={() => this.context.eventLogger.logEvent("GlobalSearchItemSelected", {globalSearchQuery: this.state.query, selectedItem: repo.URI})}>
					<div styleName="cool-gray flex-container" className={base.pt4}>
						<div styleName="flex-icon hidden-s">
							<Icon icon="repository-gray" width="32px" />
						</div>
						<div styleName="flex bottom-border" className={base.pb3}>
							<code styleName="f4 block" className={base.mb2}>
								Repository
								<span styleName="bold"> {repo.URI.split(/[// ]+/).pop()}</span>
							</code>
							<p>
								from {repo.URI}
								<span styleName="cool-mid-gray">{firstLineDocString ? ` – ${firstLineDocString}` : ""}</span>
							</p>
						</div>
					</div>
				</Link>
			);
		}

		for (let i = numRepos; i < numRepos + numDefs; i++) {
			let def = this.state.matchingResults.Defs[i - numRepos];
			let defURL = urlToDefInfo(def) ? urlToDefInfo(def) : urlToDef(def);

			const selected = this._normalizedSelectionIndex() === i;

			let docstring = "";
			if (def.Docs) {
				def.Docs.forEach((doc) => {
					if (doc.Format === "text/plain") {
						docstring = doc.Data;
					}
				});
			}

			const firstLineDocString = firstLine(docstring);
			list.push(
				<Link styleName={selected ? "block result-selected" : "block result"}
					onMouseOver={(ev) => this._mouseSelectItem(ev, i)}
					ref={selected ? this._setSelectedItem : null}
					to={defURL}
					key={defURL}
					onClick={() => this.context.eventLogger.logEvent("GlobalSearchItemSelected", {globalSearchQuery: this.state.query, selectedItem: defURL})}>
					<div styleName="cool-gray flex-container" className={base.pt4}>
						<div styleName="flex-icon hidden-s">
							<Icon icon="doc-code" width="32px" />
						</div>
						<div styleName="flex bottom-border" className={base.pb3}>
							<code styleName="f4 block" className={base.mb2}>
								{qualifiedNameAndType(def, {nameQual: "DepQualified"})}
							</code>
							<p>
								from {def.Repo}
								<span styleName="cool-mid-gray">{firstLineDocString ? ` – ${firstLineDocString}` : ""}</span>
							</p>
						</div>
					</div>
				</Link>
			);
		}

		return list;
	}

	render() {
		return (<div styleName="center flex">
			<div styleName="search-input relative">
				<Input type="text"
					block={true}
					onFocus={() => this.setState({focused: true})}
					onBlur={() => this.setState({focused: false})}
					onInput={this._handleInput}
					autoFocus={true}
					defaultValue={this.state.query}
					placeholder="Search for symbols, functions and definitions..."
					spellCheck={false}
					domRef={(e) => this._queryInput = e} />
			</div>
			<div>
				{this._results()}
			</div>
		</div>);
	}
}

export default CSSModules(GlobalSearch, styles, {allowMultiple: true});

function firstLine(text: string): string {
	text = trimLeft(text);
	let i = text.indexOf("\n");
	if (i >= 0) {
		text = text.substr(0, i);
	}
	if (text.length > 100) {
		text = text.substr(0, 100);
	}
	return text;
}
