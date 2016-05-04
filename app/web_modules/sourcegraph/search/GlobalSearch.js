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
import {urlToDef} from "sourcegraph/def/routes";
import type {Def} from "sourcegraph/def";

import {Input} from "sourcegraph/components";

import CSSModules from "react-css-modules";
import styles from "./styles/GlobalSearch.css";

export const RESULTS_LIMIT = 20;

// GlobalSearch is the global search bar + results component.
// Tech debt: this duplicates a lot of code with TreeSearch and we
// should consider merging them at some point.
class GlobalSearch extends Container {
	static contextTypes = {
		router: React.PropTypes.object.isRequired,
		siteConfig: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);

		this.state = {
			query: "",
			matchingDefs: {Defs: []},
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
		matchingDefs: {Defs: Array<Def>};
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

		state.matchingDefs = SearchStore.results.get(state.query, null, null, RESULTS_LIMIT);
	}

	onStateTransition(prevState, nextState) {
		if (prevState.query !== nextState.query) {
			Dispatcher.Backends.dispatch(new SearchActions.WantResults(nextState.query, null, null, RESULTS_LIMIT));
		}
	}

	_onChangeQuery(query: string) {
		this.context.router.replace({...this.props.location, query: {q: query || undefined}}); // eslint-disable-line no-undefined
		this.setState({query: query});
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
			this._onSelection();
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
		if (!this.state.matchingDefs || !this.state.matchingDefs.Defs) return 0;
		return Math.min(this.state.matchingDefs.Defs.length, RESULTS_LIMIT);
	}

	_normalizedSelectionIndex(): number {
		return Math.min(this.state.selectionIndex, this._numResults() - 1);
	}

	_onSelection() {
		const i = this._normalizedSelectionIndex();
		if (i === -1) {
			return;
		}
		const def = this.state.matchingDefs.Defs[i];
		this._navigateTo(urlToDef(def));
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
		if (!this.state.query) return [<div styleName="list-item list-item-empty" key="_nosymbol"></div>];

		const noResultsItem = <div styleName="list-item list-item-empty" key="_nosymbol">No results found</div>;
		if (!this.state.matchingDefs) {
			return [<div key="1" styleName="list-item list-item-empty">Loading...</div>];
		}

		if (this.state.matchingDefs && (!this.state.matchingDefs.Defs || this.state.matchingDefs.Defs.length === 0)) return [noResultsItem];

		let list = [],
			limit = this.state.matchingDefs.Defs.length > RESULTS_LIMIT ? RESULTS_LIMIT : this.state.matchingDefs.Defs.length;

		for (let i = 0; i < limit; i++) {
			let def = this.state.matchingDefs.Defs[i];
			let defURL = urlToDef(def);

			const selected = this._normalizedSelectionIndex() === i;

			let docstring = "";
			if (def.Docs) {
				def.Docs.forEach((doc) => {
					if (doc.Format === "text/plain") {
						docstring = doc.Data;
					}
				});
			}

			list.push(
				<Link styleName={selected ? "list-item-selected" : "list-item"}
					onMouseOver={(ev) => this._mouseSelectItem(ev, i)}
					ref={selected ? this._setSelectedItem : null}
					to={defURL}
					key={defURL}>
					<div styleName="search-result-main"><code>{qualifiedNameAndType(def)}</code></div>
					<div styleName="search-result-info">
							<code><span styleName="search-result-repo">{def.Repo}</span>: <span styleName="search-result-file">{def.File}</span></code><br/>
							<span styleName="search-result-ref-count">{def.RefCount}</span> examples found
					</div>
					<div styleName="search-result-doc">{firstLine(docstring)}</div>
				</Link>
			);
		}

		return list;
	}

	render() {
		return (<div styleName="container">
			<div styleName="search-section">
				<div styleName="input-container">
					<Input type="text"
						block={true}
						onFocus={() => this.setState({focused: true})}
						onBlur={() => this.setState({focused: false})}
						onInput={this._handleInput}
						autoFocus={true}
						defaultValue={this.state.query}
						placeholder="Search for symbols..."
						domRef={(e) => this._queryInput = e} />
				</div>
			</div>
			<div styleName="search-results">
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
