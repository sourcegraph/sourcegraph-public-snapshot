// @flow

import React from "react";
import ReactDOM from "react-dom";
import {Link} from "react-router";
import {browserHistory} from "react-router";
import {rel} from "sourcegraph/app/routePatterns";
import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import SearchStore from "sourcegraph/search/SearchStore";
import RepoStore from "sourcegraph/repo/RepoStore";
import "sourcegraph/search/SearchBackend";
import UserStore from "sourcegraph/user/UserStore";
import uniq from "lodash.uniq";
import debounce from "lodash.debounce";
import trimLeft from "lodash.trimleft";
import * as SearchActions from "sourcegraph/search/SearchActions";
import {qualifiedNameAndType} from "sourcegraph/def/Formatter";
import {urlToDef, urlToDefInfo} from "sourcegraph/def/routes";
import type {Options, Repo, Def} from "sourcegraph/def";
import {Icon} from "sourcegraph/components";
import {trimRepo} from "sourcegraph/repo";
import CSSModules from "react-css-modules";
import styles from "./styles/GlobalSearch.css";
import base from "sourcegraph/components/styles/_base.css";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import popularRepos from "./popularRepos";

export const RESULTS_LIMIT = 20;

const resultIconSize = "24px";

// GlobalSearch is the global search bar + results component.
// Tech debt: this duplicates a lot of code with TreeSearch and we
// should consider merging them at some point.
class GlobalSearch extends Container {
	static propTypes = {
		repo: React.PropTypes.string,
		location: React.PropTypes.object.isRequired,
		query: React.PropTypes.string.isRequired,
		className: React.PropTypes.string,
		resultClassName: React.PropTypes.string,
	};

	static contextTypes = {
		router: React.PropTypes.object.isRequired,
		eventLogger: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);

		this.state = {
			query: "",
			matchingResults: {Repos: [], Defs: [], Options: []},
			className: null,
			resultClassName: null,
			selectionIndex: -1,
			githubToken: null,
			privateRepos: [],
			publicRepos: [],
			clickModifier: false,
		};
		this._handleKeyDown = this._handleKeyDown.bind(this);
		this._handleKeyUp = this._handleKeyUp.bind(this);
		this._scrollToVisibleSelection = this._scrollToVisibleSelection.bind(this);
		this._setSelectedItem = this._setSelectedItem.bind(this);
		this._onSelection = debounce(this._onSelection.bind(this), 200, {leading: false, trailing: true});
		this._dispatchQuery = debounce(this._dispatchQuery.bind(this), 200, {leading: false, trailing: true});
	}

	state: {
		query: string;
		className: ?string;
		resultClassName: ?string;
		matchingResults: {
			Repos: Array<Repo>,
			Defs: Array<Def>,
			Options: Array<Options>,
		};
		selectionIndex: number;
		privateRepos: Array<Repo>;
		publicRepos: Array<Repo>;
		clickModifier: boolean;
	};

	componentDidMount() {
		super.componentDidMount();
		if (global.document) {
			document.addEventListener("keydown", this._handleKeyDown);
			document.addEventListener("keyup", this._handleKeyUp);
		}
		this._dispatcherToken = Dispatcher.Stores.register(this.__onDispatch.bind(this));
	}

	componentWillUnmount() {
		super.componentWillUnmount();
		if (global.document) {
			document.removeEventListener("keydown", this._handleKeyDown);
			document.removeEventListener("keyup", this._handleKeyUp);
		}
		Dispatcher.Stores.unregister(this._dispatcherToken);
	}

	_dispatcherToken: string;

	_langs(state) {
		if (!state) state = this.state;
		return state.settings && state.settings.search && state.settings.search.languages ? state.settings.search.languages : [];
	}

	_scope(state) {
		if (!state) state = this.state;
		return state.settings && state.settings.search && state.settings.search.scope ? state.settings.search.scope :
			{};
	}

	_scopeProperties(state) {
		const scope = this._scope(state);
		return Object.keys(scope).filter((key) => key === "repo" ? this.state.repo && Boolean(scope[key]) : Boolean(scope[key]));
	}

	_pageName() {
		return this.props.location.pathname.slice(1) === rel.search ? `/${rel.search}` : "(global nav)";
	}

	_canSearch(state) {
		const scope = this._scope(state);
		return scope.public || scope.private || scope.repo || (scope.popular || !state.githubToken);
	}

	_reposScope(state, lang) {
		const scope = this._scope(state);
		let repos = [];
		if (state.repo && scope.repo) repos.push(state.repo);
		if ((scope.popular || !state.githubToken) && lang) repos.push(...popularRepos[lang]);
		if (scope.public) repos.push(...state.publicRepos);
		if (scope.private) repos.push(...state.privateRepos);
		return uniq(repos);
	}

	_chunkReposScope(scope) {
		let arrays = [];
		let batchSize = 10;
		while (scope.length > 0) arrays.push(scope.splice(0, batchSize));
		return arrays;
	}

	_parseRemoteRepoURIsAndDeps(repos, deps) {
		let uris = [];
		for (let repo of repos) {
			uris.push(`github.com/${repo.Owner}/${repo.Name}`);
		}
		if (deps) uris.push(...deps.filter((dep) => dep.startsWith("github.com")));
		return uris;
	}

	stores(): Array<Object> { return [SearchStore, UserStore, RepoStore]; }

	reconcileState(state: GlobalSearch.state, props) {
		Object.assign(state, props);
		state.githubToken = UserStore.activeGitHubToken;
		state.settings = UserStore.settings.get();
		state.className = props.className || "";
		state.resultClassName = props.resultClassName || "";

		const scope = this._scope(state);
		if (scope.public) {
			const repos = RepoStore.remoteRepos.getOpt({deps: true, private: false});
			state.publicRepos = this._parseRemoteRepoURIsAndDeps(repos && repos.RemoteRepos ? repos.RemoteRepos : [], repos && repos.Dependencies ? repos.Dependencies : null);
		}
		if (scope.private) {
			const repos = RepoStore.remoteRepos.getOpt({deps: true, private: true}) || [];
			state.privateRepos = this._parseRemoteRepoURIsAndDeps(repos && repos.RemoteRepos ? repos.RemoteRepos : [], repos && repos.Dependencies ? repos.Dependencies : null);
		}

		state.matchingResults = this._langs(state).reduce((memo, lang) => {
			const reposScope = this._reposScope(state, lang);
			if (reposScope && reposScope.length > 0) {
				const batches = this._chunkReposScope(reposScope);
				for (const batch of batches) {
					const results = SearchStore.get(`${lang} ${state.query}`, batch, null, null, RESULTS_LIMIT,
						this.props.location.query.prefixMatch, this.props.location.query.includeRepos);
					if (results && !results.Error) {
						memo.fetching = false;
						if (results.Repos) memo.Repos.push(...results.Repos);
						if (results.Defs) memo.Defs.push(...results.Defs);
						if (results.Options) memo.Options.push(...results.Options);
					}
				}
			}
			return memo;
		}, {Repos: [], Defs: [], Options: [], fetching: true});
	}

	onStateTransition(prevState, nextState) {
		if (prevState.query !== nextState.query ||
			prevState.githubToken !== nextState.githubToken ||
			prevState.settings !== nextState.settings ||
			prevState.publicRepos !== nextState.publicRepos ||
			prevState.privateRepos !== nextState.privateRepos) {
			this._dispatchQuery(nextState);
		}
	}

	_dispatchQuery(state) {
		const langs = this._langs(state);
		for (const lang of langs) {
			const reposScope = this._reposScope(state, lang);
			if (!reposScope || reposScope.length === 0 || !this._canSearch(state)) continue;
			const batches = this._chunkReposScope(reposScope);
			for (const batch of batches) {
				Dispatcher.Backends.dispatch(new SearchActions.WantResults({
					query: `${lang} ${state.query}`,
					limit: RESULTS_LIMIT,
					prefixMatch: this.props.location.query.prefixMatch,
					includeRepos: this.props.location.query.includeRepos,
					fast: true,
					repos: batch,
				}));
			}
		}
	}

	__onDispatch(action) {
		if (action.constructor === SearchActions.ResultsFetched) {
			let eventProps = {};
			eventProps["globalSearchQuery"] = this.state.query;
			eventProps["page name"] = this._pageName();
			eventProps["languages"] = this._langs(this.state);
			eventProps["repo_scope"] = this._scopeProperties(this.state);
			this.context.eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_GLOBAL_SEARCH, AnalyticsConstants.ACTION_SUCCESS, "GlobalSearchInitiated", eventProps);
		}
	}

	_navigateTo(url: string) {
		if (!this.state.clickModifier) {
			browserHistory.push(url);
		}
	}

	_handleKeyDown(e: KeyboardEvent) {
		let code = e.keyCode;
		if (e.metaKey) {
			// http://stackoverflow.com/questions/3902635/how-does-one-capture-a-macs-command-key-via-javascript
			code = 17;
		}
		let idx, max;
		switch (code) {
		case 40: // ArrowDown
			idx = this._normalizedSelectionIndex();
			max = this._numResults();

			this.setState({
				selectionIndex: idx + 1 >= max ? -1 : idx + 1,
			}, this._scrollToVisibleSelection);

			this._temporarilyIgnoreMouseSelection();
			e.preventDefault();
			break;

		case 38: // ArrowUp
			idx = this._normalizedSelectionIndex();
			max = this._numResults();

			this.setState({
				selectionIndex: idx < 0 ? max-1 : idx-1,
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
			// Ignore global search enter keypress (to submit search form).
			if (this._normalizedSelectionIndex() !== -1) {
				this._onSelection();
				this._temporarilyIgnoreMouseSelection();
				e.preventDefault();
			}
			break;

		case 17: // ctrl (or meta, see above)
			this.setState({clickModifier: true});
			break;

		default:
			// Changes to the input value are handled by the parent component.
			break;
		}
	}

	_handleKeyUp(e: KeyboardEvent) {
		this.setState({clickModifier: false});
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

		let eventProps: any = {globalSearchQuery: this.state.query, indexSelected: i, page_name: this._pageName(), languages: this._langs(this.state), repo_scope: this._scopeProperties(this.state)};

		let offset = 0;
		if (this.state.matchingResults.Repos) {
			if (i < this.state.matchingResults.Repos.length) {
				const url = `/${this.state.matchingResults.Repos[i].URI}`;
				eventProps.selectedItem = url;
				this.context.eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_GLOBAL_SEARCH, AnalyticsConstants.ACTION_CLICK, "GlobalSearchItemSelected", eventProps);
				this._navigateTo(url);
				return;
			}

			offset = this.state.matchingResults.Repos.length;
		}

		const def = this.state.matchingResults.Defs[i - offset];
		const url = urlToDefInfo(def) ? urlToDefInfo(def) : urlToDef(def);

		eventProps.selectedItem = url;
		eventProps.totalResults = this.state.matchingResults.Defs.length;
		if (def.FmtStrings && def.FmtStrings.Kind && def.FmtStrings.Language && def.Repo) {
			eventProps = {...eventProps, languageSelected: def.FmtStrings.Language, kindSelected: def.FmtStrings.Kind, repoSelected: def.Repo};
		}

		this.context.eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_GLOBAL_SEARCH, AnalyticsConstants.ACTION_CLICK, "GlobalSearchItemSelected", eventProps);

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

	_results(): React$Element | Array<React$Element> {
		const langs = this._langs(this.state);
		const scope = this._scope(this.state);

		if (!langs || langs.length === 0) {
			return [<div key="_nosymbol" className={`${base.ph4} ${base.pt4}`} styleName="result result-error">Select a language to search.</div>];
		}

		if (!scope || !(scope.popular || (this.state.repo && scope.repo) || scope.private || scope.public)) {
			return [<div key="_nosymbol" className={`${base.ph4} ${base.pt4}`} styleName="result result-error">Select repositories to include.</div>];
		}

		if (!this.state.query) return <div className={`${base.pt4} ${base.ph4}`} styleName="result">Type a query&hellip;</div>;

		if (!this.state.matchingResults || this.state.matchingResults && this.state.matchingResults.fetching) {
			return [<div key="_nosymbol" className={`${base.ph4} ${base.pt4}`}styleName="result">Loading results...</div>];
		}

		if (this.state.matchingResults &&
			(!this.state.matchingResults.Defs || this.state.matchingResults.Defs.length === 0) &&
			(!this.state.matchingResults.Repos || this.state.matchingResults.Repos.length === 0)) {
			return [<div className={`${base.ph4} ${base.pt4}`} styleName="result" key="_nosymbol">No results found.</div>];
		}

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
					className={this.state.resultClassName}
					onMouseOver={(ev) => this._mouseSelectItem(ev, i)}
					ref={selected ? this._setSelectedItem : null}
					to={repo.URI}
					key={repo.URI}
					onClick={() => this._onSelection()}>
					<div styleName="cool-gray flex-container">
						<div styleName="flex-icon hidden-s">
							<Icon icon="repository-gray" width={resultIconSize} />
						</div>
						<div styleName="flex">
							<code styleName="block f5">
								Repository
								<span styleName="bold"> {repo.URI.split(/[// ]+/).pop()}</span>
							</code>
							{firstLineDocString && <p styleName="cool-mid-gray" className={base.mt0}>{firstLineDocString}</p>}
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
					className={this.state.resultClassName}
					onMouseOver={(ev) => this._mouseSelectItem(ev, i)}
					ref={selected ? this._setSelectedItem : null}
					to={defURL}
					key={defURL}
					onClick={() => this._onSelection()}>
					<div styleName="cool-gray flex-container" className={base.pt3}>
						<div styleName="flex">
							<p styleName="cool-mid-gray block-s" className={`${base.ma0} ${base.pl4} ${base.pr2} ${base.fr}`}>{trimRepo(def.Repo)}</p>
							<code styleName="block f5" className={base.pb3}>
								{qualifiedNameAndType(def, {nameQual: "DepQualified"})}
							</code>
							{firstLineDocString && <p styleName="cool-mid-gray" className={base.mt0}>{firstLineDocString}</p>}
						</div>
					</div>
				</Link>
			);
		}

		return list;
	}

	render() {
		return (<div styleName="center flex" className={this.state.className}>
			{this._results()}
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
