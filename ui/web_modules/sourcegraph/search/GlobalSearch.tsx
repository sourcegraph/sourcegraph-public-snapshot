// tslint:disable: typedef ordered-imports

import * as React from "react";
import * as ReactDOM from "react-dom";
import {Link} from "react-router";
import {rel} from "sourcegraph/app/routePatterns";
import {Container} from "sourcegraph/Container";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {SearchStore} from "sourcegraph/search/SearchStore";
import {RepoStore} from "sourcegraph/repo/RepoStore";
import "sourcegraph/search/SearchBackend";
import {UserStore} from "sourcegraph/user/UserStore";
import uniq from "lodash.uniq";
import debounce from "lodash.debounce";
import * as SearchActions from "sourcegraph/search/SearchActions";
import {qualifiedNameAndType} from "sourcegraph/def/Formatter";
import {urlToDef, urlToDefInfo} from "sourcegraph/def/routes";
import {Options, Repo, Def} from "sourcegraph/def/index";
import {Icon} from "sourcegraph/components/index";
import {trimRepo} from "sourcegraph/repo/index";
import * as styles from "./styles/GlobalSearch.css";
import * as base from "sourcegraph/components/styles/_base.css";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {popularRepos} from "./popularRepos";
import {SearchSettings} from "sourcegraph/search/index";
import {WantResultsPayload} from "sourcegraph/search/SearchActions";
import {locationForSearch} from "sourcegraph/search/routes";
import * as classNames from "classnames";

export const RESULTS_LIMIT = 20;

const resultIconSize = "24px";

interface Props {
	repo: string | null;
	location: any;
	query: string;
	className?: string;
	resultClassName?: string;
}

// GlobalSearch is the global search bar + results component.
export class GlobalSearch extends Container<Props, any> {
	static contextTypes = {
		router: React.PropTypes.object.isRequired,
		eventLogger: React.PropTypes.object.isRequired,
	};

	_selectedItem: any;
	_ignoreMouseSelection: any;
	_debouncedUnignoreMouseSelection: any;
	_dispatcherToken: string;
	_debounceForSearch = debounce((f: Function) => f(), 200, {leading: false, trailing: true});

	state: {
		repo: string | null;
		query: string;
		className: string | null;
		resultClassName: string | null;
		matchingResults: {
			Repos: Array<Repo>,
			Defs: Array<Def>,
			Options: Array<Options>,
			outstandingFetches: number,
		};
		selectionIndex: number;
		githubToken: any;

		searchSettings: SearchSettings | null;

		_queries: Array<WantResultsPayload> | null;
		_searchStore: any,
		_privateRepos: Array<Repo>;
		_publicRepos: Array<Repo>;
		_reposByLang: any;
	};

	constructor(props: Props) {
		super(props);

		this.state = {
			query: "",
			repo: null,
			matchingResults: {Repos: [], Defs: [], Options: [], outstandingFetches: 0},
			className: null,
			resultClassName: null,
			selectionIndex: 0,
			githubToken: null,
			searchSettings: null,
			_queries: null,
			_searchStore: null,
			_privateRepos: [],
			_publicRepos: [],
			_reposByLang: null,
		};
		this._handleKeyDown = this._handleKeyDown.bind(this);
		this._scrollToVisibleSelection = this._scrollToVisibleSelection.bind(this);
		this._setSelectedItem = this._setSelectedItem.bind(this);
		this._onSelection = debounce(this._onSelection.bind(this), 200, {leading: false, trailing: true});
	}

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

	_scopeProperties(): string[] | null {
		const scope = this.state.searchSettings ? this.state.searchSettings.scope : null;
		if (!scope) {
			return null;
		}
		return Object.keys(scope).filter((key) => key === "repo" ? this.state.repo && Boolean(scope[key]) : Boolean(scope[key]));
	}

	_pageName() {
		return this.props.location.pathname.slice(1) === rel.search ? `/${rel.search}` : "(global nav)";
	}

	_parseRemoteRepoURIsAndDeps(repos, deps) {
		let uris: any[] = [];
		for (let repo of repos) {
			uris.push(`github.com/${repo.Owner}/${repo.Name}`);
		}
		if (deps) {
			uris.push(...deps.filter((dep) => dep.startsWith("github.com")));
		}
		return uris;
	}

	stores(): FluxUtils.Store<any>[] { return [SearchStore, UserStore, RepoStore]; }

	reconcileState(state, props: Props) {
		Object.assign(state, props);
		state.githubToken = UserStore.activeGitHubToken;
		state.language = state.searchSettings && state.searchSettings.languages ? state.searchSettings.languages : null;
		state.className = props.className || "";
		state.resultClassName = props.resultClassName || "";

		const settings = UserStore.settings;
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
					const repos: any[] = [];
					if (state.repo && scope.repo) {
						repos.push(state.repo);
					}
					if (scope.popular && lang) {
						repos.push(...popularRepos[lang]);
					}
					if (scope.public) {
						repos.push(...state._publicRepos);
					}
					if (scope.private) {
						repos.push(...state._privateRepos);
					}
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
					const results = SearchStore.get(q.query, q.repos, q.notRepos, q.commitID, q.limit);
					if (results) {
						memo.outstandingFetches -= 1;
					}
					if (results && !results.Error) {
						if (results.Repos) {
							memo.Repos.push(...results.Repos);
						}
						if (results.Defs) {
							memo.Defs.push(...results.Defs);
						}
						if (results.Options) {
							memo.Options.push(...results.Options);
						}
					}
					return memo;
				}, {Repos: [], Defs: [], Options: [], outstandingFetches: state._queries.length});
			} else {
				state.matchingResults = null;
			}
		}
	}

	onStateTransition(prevState, nextState) {
		if (prevState.searchSettings && prevState.searchSettings !== nextState.searchSettings && nextState.location.pathname === "/search") {
			(this.context as any).router.replace(locationForSearch(nextState.location, nextState.query, nextState.searchSettings.languages, nextState.searchSettings.scope, false, true));
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
			(this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_GLOBAL_SEARCH, AnalyticsConstants.ACTION_SUCCESS, "GlobalSearchInitiated", eventProps);
		}
	}

	_navigateTo(url: string) {
		(this.context as any).router.push(url);
	}

	_handleKeyDown(e: KeyboardEvent) {
		let idx;
		let max;
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
				selectionIndex: idx <= 0 ? max - 1 : idx - 1,
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
			this._onSelection(false);
			this._temporarilyIgnoreMouseSelection();
			e.preventDefault();
			break;
		default:
			// Changes to the input value are handled by the parent component.
			break;
		}
	}

	_scrollToVisibleSelection() {
		if (this._selectedItem) {
			(ReactDOM.findDOMNode(this._selectedItem) as HTMLElement).scrollIntoView(false);
		}
	}

	_setSelectedItem(e: any) {
		this._selectedItem = e;
	}

	_numResults(): number {
		if (!this.state.matchingResults ||
			(!this.state.matchingResults.Defs && !this.state.matchingResults.Repos)) {
				return 0;
			}

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
	_onSelection(trackOnly: boolean) {
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
				(this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_GLOBAL_SEARCH, AnalyticsConstants.ACTION_CLICK, "GlobalSearchItemSelected", eventProps);
				if (!trackOnly) {
					this._navigateTo(url);
				}
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
			eventProps = Object.assign({}, eventProps, {languageSelected: def.FmtStrings.Language, kindSelected: def.FmtStrings.Kind, repoSelected: def.Repo});
		}

		(this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_GLOBAL_SEARCH, AnalyticsConstants.ACTION_CLICK, "GlobalSearchItemSelected", eventProps);

		if (!trackOnly) {
			this._navigateTo(url);
		}
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
	_mouseSelectItem(ev: React.MouseEvent, i: number): void {
		if (this._ignoreMouseSelection) {
			return;
		}
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

	_results() {
		const langs = this.state.searchSettings ? this.state.searchSettings.languages : null;

		if (!langs || langs.length === 0) {
			return [<div key="_nosymbol" className={classNames(base.ph4, base.pt4, styles.result, styles.result_error)}>Select a language to search.</div>];
		}

		if (this.state.query && !this.state.matchingResults ||
			((!this.state.matchingResults.Defs || this.state.matchingResults.Defs.length === 0) && this.state.matchingResults.outstandingFetches !== 0) && this.state.query) {
			return [<div key="_nosymbol" className={classNames(base.ph4, base.pv4, styles.result)}>Loading results...</div>];
		}

		if (this.state.query && this.state.matchingResults &&
			(!this.state.matchingResults.Defs || this.state.matchingResults.Defs.length === 0) &&
			(!this.state.matchingResults.Repos || this.state.matchingResults.Repos.length === 0)) {
			return [<div className={classNames(base.ph4, base.pv4, styles.result)} key="_nosymbol">No results found.</div>];
		}

		let list: any[] = [];
		let numDefs = 0;
		let numRepos = this.state.matchingResults.Repos ? this.state.matchingResults.Repos.length : 0;

		if (this.state.matchingResults.Defs) {
			numDefs = this.state.matchingResults.Defs.length > RESULTS_LIMIT ? RESULTS_LIMIT : this.state.matchingResults.Defs.length;
		}
		for (let i = 0; i < numRepos; i++) {
			let repo = this.state.matchingResults.Repos[i];
			const selected = this._normalizedSelectionIndex() === i;

			const firstLineDocString = repo.Description;
			list.push(
				<Link className={classNames(styles.block, selected ? styles.result_selected : styles.result, this.state.resultClassName)}
					onMouseOver={(ev) => this._mouseSelectItem(ev, i)}
					ref={selected ? this._setSelectedItem : null}
					to={repo.URI}
					key={repo.URI}
					onClick={() => this._onSelection(true)}>
					<div className={classNames(styles.cool_gray, styles.flex_container)}>
						<div className={classNames(styles.flex_icon, styles.hidden_s)}>
							<Icon icon="repository-gray" width={resultIconSize} />
						</div>
						<div className={styles.flex}>
							<code className={classNames(styles.block, styles.f5)}>
								Repository
								<span className={styles.bold}> {repo.URI.split(/[// ]+/).pop()}</span>
							</code>
							{firstLineDocString && <p className={classNames(styles.docstring, base.mt0)}>{firstLineDocString}</p>}
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
				<Link className={classNames(styles.block, selected ? styles.result_selected : styles.result, this.state.resultClassName)}
					onMouseOver={(ev) => this._mouseSelectItem(ev, i)}
					ref={selected ? this._setSelectedItem : null}
					to={defURL.replace(/GoPackage\/pkg\//, "GoPackage/")}
					key={defURL}
					onClick={() => this._onSelection(true)}>
					<div className={classNames(styles.cool_gray, styles.flex_container, base.pt3)}>
						<div className={classNames(styles.flex, styles.w100)}>
							<p className={classNames(styles.cool_mid_gray, styles.block_s, base.ma0, base.pl4, base.pr2, base.fr)}>{trimRepo(def.Repo)}</p>
							<code className={classNames(styles.block, styles.f5, base.pb3)}>
								{qualifiedNameAndType(def, {nameQual: "DepQualified"})}
							</code>
							{firstLineDocString && <p className={classNames(styles.docstring, base.mt0)}>{firstLineDocString}</p>}
						</div>
					</div>
				</Link>
			);
		}

		return list;
	}

	render(): JSX.Element | null {
		return (<div className={classNames(styles.center, styles.flex, this.state.className)}>
			{this._results()}
		</div>);
	}
}
