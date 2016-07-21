// @flow

import React from "react";
import ReactDOM from "react-dom";
import {Link} from "react-router";
import {rel} from "sourcegraph/app/routePatterns";
import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import SearchStore from "sourcegraph/search/SearchStore";
import RepoStore from "sourcegraph/repo/RepoStore";
import "sourcegraph/search/SearchBackend";
import UserStore from "sourcegraph/user/UserStore";
import uniq from "lodash.uniq";
import debounce from "lodash.debounce";
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
import type {SearchSettings} from "sourcegraph/search";
import type {WantResultsPayload} from "sourcegraph/search/SearchActions";
import {locationForSearch} from "sourcegraph/search/routes";

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
			repo: null,
			matchingResults: {Repos: [], Defs: [], Options: [], outstandingFetches: 0},
			className: null,
			resultClassName: null,
			selectionIndex: -1,
			githubToken: null,
			searchSettings: null,
			_queries: null,
			_searchStore: null,
			_privateRepos: [],
			_publicRepos: [],
		};
		this._handleKeyDown = this._handleKeyDown.bind(this);
		this._scrollToVisibleSelection = this._scrollToVisibleSelection.bind(this);
		this._setSelectedItem = this._setSelectedItem.bind(this);
		this._onSelection = debounce(this._onSelection.bind(this), 200, {leading: false, trailing: true});
	}

	state: {
		repo: ?string;
		query: string;
		className: ?string;
		resultClassName: ?string;
		matchingResults: {
			Repos: Array<Repo>,
			Defs: Array<Def>,
			Options: Array<Options>,
			outstandingFetches: number,
		};
		selectionIndex: number;

		searchSettings: ?SearchSettings;

		_queries: ?Array<WantResultsPayload>;
		_searchStore: ?Object,
		_privateRepos: Array<Repo>;
		_publicRepos: Array<Repo>;
	};

	componentDidMount() {
		super.componentDidMount();
		if (global.document) {
			document.addEventListener("keydown", this._handleKeyDown);
		}
		this._dispatcherToken = Dispatcher.Stores.register(this.__onDispatch.bind(this));
	}

	componentWillUnmount() {
		super.componentWillUnmount();
		if (global.document) {
			document.removeEventListener("keydown", this._handleKeyDown);
		}
		Dispatcher.Stores.unregister(this._dispatcherToken);
	}

	_dispatcherToken: string;

	_scopeProperties(): ?string[] {
		const scope = this.state.searchSettings ? this.state.searchSettings.scope : null;
		if (!scope) return null;
		return Object.keys(scope).filter((key) => key === "repo" ? this.state.repo && Boolean(scope[key]) : Boolean(scope[key]));
	}

	_pageName() {
		return this.props.location.pathname.slice(1) === rel.search ? `/${rel.search}` : "(global nav)";
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
		state.language = state.searchSettings && state.searchSettings.languages ? state.searchSettings.languages : null;
		state.className = props.className || "";
		state.resultClassName = props.resultClassName || "";

		const settings = UserStore.settings.get();
		state.searchSettings = settings && settings.search ? settings.search : null;
		const scope = state.searchSettings && state.searchSettings.scope ? state.searchSettings.scope : null;
		const languages = state.searchSettings && state.searchSettings.languages ? state.searchSettings.languages : null;
		if (this.state.searchSettings !== state.searchSettings) {
			if (scope && scope.public) {
				const repos = RepoStore.repos.list("Private=false");
				state._publicRepos = this._parseRemoteRepoURIsAndDeps(repos && repos.Repos ? repos.Repos : [], repos && repos.Dependencies ? repos.Dependencies : null);
			} else {
				state._publicRepos = null;
			}
			if (scope && scope.private) {
				const repos = RepoStore.repos.list("Private=true") || [];
				state._privateRepos = this._parseRemoteRepoURIsAndDeps(repos && repos.Repos ? repos.Repos : [], repos && repos.Dependencies ? repos.Dependencies : null);
			} else {
				state._privateRepos = null;
			}

		}

		if (this.state.repo !== state.repo || this.state.searchSettings !== state.searchSettings || this.state._publicRepos !== state._publicRepos || this.state._privateRepos !== state._privateRepos) {
			if (languages && scope) {
				state._reposByLang = {};
				for (const lang of languages) {
					const repos = [];
					if (state.repo && scope.repo) repos.push(state.repo);
					if (scope.popular && lang) repos.push(...popularRepos[lang]);
					if (scope.public) repos.push(...state._publicRepos);
					if (scope.private) repos.push(...state._privateRepos);
					state._reposByLang[lang] = uniq(repos);
				}
			} else {
				state._reposByLang = null;
			}
		}

		if (this.state.searchSettings !== state.searchSettings || this.state.query !== state.query || this.state._reposByLang !== state._reposByLang) {
			if (languages && state._reposByLang) {
				state._queries = [];
				for (const lang of languages) {
					const repos = state._reposByLang[lang];
					state._queries.push({
						query: `${lang} ${state.query}`,
						repos: repos,
						limit: RESULTS_LIMIT,
						includeRepos: props.location.query.includeRepos,
						fast: true,
					});
				}
			} else {
				state._queries = null;
			}
		}

		state._searchStore = SearchStore.content;
		if (this.state._searchStore !== state._searchStore || this.state._queries !== state._queries) {
			if (state._queries) {
				state.matchingResults = state._queries.reduce((memo, q) => {
					const results = SearchStore.get(q.query, q.repos, q.notRepos, q.commitID, q.limit, q.includeRepos, q.fast);
					if (results) memo.outstandingFetches -= 1;
					if (results && !results.Error) {
						if (results.Repos) memo.Repos.push(...results.Repos);
						if (results.Defs) memo.Defs.push(...results.Defs);
						if (results.Options) memo.Options.push(...results.Options);
					}
					return memo;
				}, {Repos: [], Defs: [], Options: [], outstandingFetches: state._queries.length});
			} else {
				state.matchingResults = null;
			}
		}
	}

	_debounceForSearch = debounce((f: Function) => f(), 200, {leading: false, trailing: true});

	onStateTransition(prevState, nextState) {
		if (prevState.searchSettings && prevState.searchSettings !== nextState.searchSettings && nextState.location.pathname === "/search") {
			this.context.router.replace(locationForSearch(nextState.location, nextState.query, nextState.searchSettings.languages, nextState.searchSettings.scope, false, true));
		}

		if (prevState.githubToken !== nextState.githubToken ||
			prevState._queries !== nextState._queries) {
			if (nextState._queries) {
				this._debounceForSearch(() => {
					for (const q of nextState._queries) {
						Dispatcher.Backends.dispatch(new SearchActions.WantResults(q));
					}
				});
			}
		}
	}

	__onDispatch(action) {
		if (action instanceof SearchActions.ResultsFetched) {
			let eventProps = {};
			eventProps["globalSearchQuery"] = this.state.query;
			eventProps["page name"] = this._pageName();
			eventProps["languages"] = this.state.searchSettings ? this.state.searchSettings.languages : null;
			eventProps["repo_scope"] = this._scopeProperties();
			this.context.eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_GLOBAL_SEARCH, AnalyticsConstants.ACTION_SUCCESS, "GlobalSearchInitiated", eventProps);
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
				this._onSelection(false);
				this._temporarilyIgnoreMouseSelection();
				e.preventDefault();
			}
			break;
		default:
			// Changes to the input value are handled by the parent component.
			break;
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

	// _onSelection handles a selection of a result. The trackOnly param means that the
	// result should not actually be navigated to.
	_onSelection(trackOnly: bool) {
		const i = this._normalizedSelectionIndex();
		if (i === -1) {
			return;
		}

		let eventProps: any = {
			globalSearchQuery: this.state.query,
			indexSelected: i,
			page_name: this._pageName(),
			languages: this.state.searchSettings ? this.state.searchSettings.languages : null,
			repo_scope: this._scopeProperties(),
		};

		let offset = 0;
		if (this.state.matchingResults.Repos) {
			if (i < this.state.matchingResults.Repos.length) {
				const url = `/${this.state.matchingResults.Repos[i].URI}`;
				eventProps.selectedItem = url;
				this.context.eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_GLOBAL_SEARCH, AnalyticsConstants.ACTION_CLICK, "GlobalSearchItemSelected", eventProps);
				if (!trackOnly) this._navigateTo(url);
				return;
			}

			offset = this.state.matchingResults.Repos.length;
		}

		const def = this.state.matchingResults.Defs[i - offset];
		let url = urlToDefInfo(def) ? urlToDefInfo(def) : urlToDef(def);
		url = url.replace(/GoPackage\/pkg\//, "GoPackage/"); // TEMP HOTFIX

		eventProps.selectedItem = url;
		eventProps.totalResults = this.state.matchingResults.Defs.length;
		if (def.FmtStrings && def.FmtStrings.Kind && def.FmtStrings.Language && def.Repo) {
			eventProps = {...eventProps, languageSelected: def.FmtStrings.Language, kindSelected: def.FmtStrings.Kind, repoSelected: def.Repo};
		}

		this.context.eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_GLOBAL_SEARCH, AnalyticsConstants.ACTION_CLICK, "GlobalSearchItemSelected", eventProps);

		if (!trackOnly) this._navigateTo(url);
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
		const langs = this.state.searchSettings ? this.state.searchSettings.languages : null;

		if (!langs || langs.length === 0) {
			return [<div key="_nosymbol" className={`${base.ph4} ${base.pt4}`} styleName="result result-error">Select a language to search.</div>];
		}

		if (this.state.query && !this.state.matchingResults ||
			((!this.state.matchingResults.Defs || this.state.matchingResults.Defs.length === 0) && this.state.matchingResults.outstandingFetches !== 0) && this.state.query) {
			return [<div key="_nosymbol" className={`${base.ph4} ${base.pv4}`} styleName="result">Loading results...</div>];
		}

		if (this.state.query && this.state.matchingResults &&
			(!this.state.matchingResults.Defs || this.state.matchingResults.Defs.length === 0) &&
			(!this.state.matchingResults.Repos || this.state.matchingResults.Repos.length === 0)) {
			return [<div className={`${base.ph4} ${base.pv4}`} styleName="result" key="_nosymbol">No results found.</div>];
		}

		let list = [], numDefs = 0,
			numRepos = this.state.matchingResults.Repos ? this.state.matchingResults.Repos.length : 0;

		if (this.state.matchingResults.Defs) {
			numDefs = this.state.matchingResults.Defs.length > RESULTS_LIMIT ? RESULTS_LIMIT : this.state.matchingResults.Defs.length;
		}
		for (let i = 0; i < numRepos; i++) {
			let repo = this.state.matchingResults.Repos[i];
			const selected = this._normalizedSelectionIndex() === i;

			const firstLineDocString = repo.Description;
			list.push(
				<Link styleName={selected ? "block result-selected" : "block result"}
					className={this.state.resultClassName}
					onMouseOver={(ev) => this._mouseSelectItem(ev, i)}
					ref={selected ? this._setSelectedItem : null}
					to={repo.URI}
					key={repo.URI}
					onClick={() => this._onSelection(true)}>
					<div styleName="cool-gray flex-container">
						<div styleName="flex-icon hidden-s">
							<Icon icon="repository-gray" width={resultIconSize} />
						</div>
						<div styleName="flex">
							<code styleName="block f5">
								Repository
								<span styleName="bold"> {repo.URI.split(/[// ]+/).pop()}</span>
							</code>
							{firstLineDocString && <p styleName="docstring" className={base.mt0}>{firstLineDocString}</p>}
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

			const firstLineDocString = docstring;
			list.push(
				<Link styleName={selected ? "block result-selected" : "block result"}
					className={this.state.resultClassName}
					onMouseOver={(ev) => this._mouseSelectItem(ev, i)}
					ref={selected ? this._setSelectedItem : null}
					to={defURL.replace(/GoPackage\/pkg\//, "GoPackage/")}
					key={defURL}
					onClick={() => this._onSelection(true)}>
					<div styleName="cool-gray flex-container" className={base.pt3}>
						<div styleName="flex w100">
							<p styleName="cool-mid-gray block-s" className={`${base.ma0} ${base.pl4} ${base.pr2} ${base.fr}`}>{trimRepo(def.Repo)}</p>
							<code styleName="block f5" className={base.pb3}>
								{qualifiedNameAndType(def, {nameQual: "DepQualified"})}
							</code>
							{firstLineDocString && <p styleName="docstring" className={base.mt0}>{firstLineDocString}</p>}
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
