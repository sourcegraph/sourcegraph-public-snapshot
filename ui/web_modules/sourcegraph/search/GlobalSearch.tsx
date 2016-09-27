import * as classNames from "classnames";
import * as debounce from "lodash/debounce";
import * as uniq from "lodash/uniq";
import * as React from "react";
import * as ReactDOM from "react-dom";
import {InjectedRouter, Link} from "react-router";
import {context} from "sourcegraph/app/context";
import {EventListener} from "sourcegraph/Component";
import {Icon} from "sourcegraph/components";
import * as base from "sourcegraph/components/styles/_base.css";
import {Container} from "sourcegraph/Container";
import {urlToDefInfo} from "sourcegraph/def/routes";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {trimRepo} from "sourcegraph/repo";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import {RepoStore} from "sourcegraph/repo/RepoStore";
import {urlToRepo} from "sourcegraph/repo/routes";
import {popularRepos} from "sourcegraph/search/popularRepos";
import * as SearchActions from "sourcegraph/search/SearchActions";
import "sourcegraph/search/SearchBackend";
import {SearchStore} from "sourcegraph/search/SearchStore";
import * as styles from "sourcegraph/search/styles/GlobalSearch.css";
import {Store} from "sourcegraph/Store";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {EventLogger} from "sourcegraph/util/EventLogger";

export const RESULTS_LIMIT = 20;

const resultIconSize = "24px";

interface Props {
	location: Location;
	query: string;
	className: string;
	resultClassName: string;
}

interface State {
	query: string;
	className: string;
	resultClassName: string;
	matchingResults: any;
	selectionIndex: number;

	location: Location;
	languages: string[];

	_queries: any;
	_searchStore: any;
	_privateRepos: any;
	_publicRepos: any;
	_reposByLang: any;

	_shouldFetchPublicRepos: boolean;
	_shouldFetchPrivateRepos: boolean;
}

// GlobalSearch is the global search bar + results component.
export class GlobalSearch extends Container<Props, State> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: {router: InjectedRouter};
	_selectedItem: any;
	_ignoreMouseSelection: any;
	_debouncedUnignoreMouseSelection: any;
	_dispatcherToken: string;
	_debounceForSearch: any = debounce((f: Function) => f(), 200, { leading: false, trailing: true });

	constructor(props: Props) {
		super(props);

		this.state = {
			query: "",
			matchingResults: { Repos: [], Defs: [], Options: [], outstandingFetches: 0 },
			className: props.className,
			resultClassName: props.resultClassName,
			selectionIndex: 0,
			location: props.location,
			languages: ["golang", "java"],
			_queries: null,
			_searchStore: null,
			_privateRepos: [],
			_publicRepos: [],
			_reposByLang: null,

			// These bits control whether we should fetch the current user's public/private repos list.
			_shouldFetchPublicRepos: false,
			_shouldFetchPrivateRepos: false,
		};
		this._handleKeyDown = this._handleKeyDown.bind(this);
		this._scrollToVisibleSelection = this._scrollToVisibleSelection.bind(this);
		this._setSelectedItem = this._setSelectedItem.bind(this);
		this._onSelection = debounce(this._onSelection.bind(this), 200, { leading: false, trailing: true });
	}

	componentDidMount(): void {
		super.componentDidMount();
		this._dispatcherToken = Dispatcher.Stores.register(this.__onDispatch.bind(this));
	}

	componentWillUnmount(): void {
		super.componentWillUnmount();
		Dispatcher.Stores.unregister(this._dispatcherToken);
	}

	_parseRemoteRepoURIsAndDeps(repos: any, deps: any): any[] {
		let uris: any[] = [];
		for (let repo of repos) {
			uris.push(`github.com/${repo.Owner}/${repo.Name}`);
		}
		if (deps) {
			uris.push(...deps.filter((dep) => dep.startsWith("github.com")));
		}
		return uris;
	}

	stores(): Store<any>[] {
		return [SearchStore, RepoStore];
	}

	reconcileState(state: State, props: Props): void {
		Object.assign(state, props);

		state.className = props.className || "";
		state.resultClassName = props.resultClassName || "";

		const scope = {
			public: Boolean(context.user),
			private: context.hasPrivateGitHubToken(),
			popular: true,
		};

		if (scope.public) {
			const repos = RepoStore.repos.list("RemoteOnly=true&Type=public") || [];
			state._publicRepos = this._parseRemoteRepoURIsAndDeps(repos && repos.Repos ? repos.Repos : [], repos && repos.Dependencies ? repos.Dependencies : null);
		} else {
			state._publicRepos = null;
		}
		if (scope.private) {
			const repos = RepoStore.repos.list("RemoteOnly=true&Type=private") || [];
			state._privateRepos = this._parseRemoteRepoURIsAndDeps(repos && repos.Repos ? repos.Repos : [], repos && repos.Dependencies ? repos.Dependencies : null);
		} else {
			state._privateRepos = null;
		}

		if (this.state._publicRepos !== state._publicRepos || this.state._privateRepos !== state._privateRepos) {
			state._reposByLang = {};
			for (const lang of state.languages) {
				const repos: any[] = [];
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
		}

		if (this.state.query !== state.query || this.state._reposByLang !== state._reposByLang) {
			if (state.languages && state._reposByLang) {
				state._queries = [];
				for (const lang of state.languages) {
					const repos = state._reposByLang[lang];
					state._queries.push({
						query: `${lang} ${state.query}`,
						repos: repos,
						limit: RESULTS_LIMIT,
						includeRepos: false,
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
				}, { Repos: [], Defs: [], Options: [], outstandingFetches: state._queries.length });
			} else {
				state.matchingResults = null;
			}
		}

		state._shouldFetchPublicRepos = scope.public;
		state._shouldFetchPrivateRepos = scope.private;
	}

	onStateTransition(prevState: State, nextState: State): void {
		if (prevState._queries !== nextState._queries) {
			if (nextState._queries) {
				this._debounceForSearch(() => {
					for (const q of nextState._queries) {
						Dispatcher.Backends.dispatch(new SearchActions.WantResults(q));
					}
				});
			}
		}

		// Fetch the user's repos, so we can provide search results for their repos.
		if (!prevState._shouldFetchPublicRepos && nextState._shouldFetchPublicRepos) {
			Dispatcher.Backends.dispatch(new RepoActions.WantRepos("RemoteOnly=true&Type=public"));
		}
		if (!prevState._shouldFetchPrivateRepos && nextState._shouldFetchPrivateRepos) {
			Dispatcher.Backends.dispatch(new RepoActions.WantRepos("RemoteOnly=true&Type=private"));
		}
	}

	__onDispatch(action: any): void {
		if (action instanceof SearchActions.ResultsFetched) {
			let eventProps = {};
			eventProps["globalSearchQuery"] = this.state.query;
			EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_GLOBAL_SEARCH, AnalyticsConstants.ACTION_SUCCESS, "GlobalSearchInitiated", eventProps);
		}
	}

	_navigateTo(url: string): void {
		if (url.indexOf("/info/") !== -1) {
			// The def landing page is rendered on the server and we
			// must initiate a full refresh to display it.
			window.location.href = url;
		} else {
			this.context.router.push(url);
		}
	}

	_handleKeyDown(e: KeyboardEvent): void {
		let idx;
		let max;
		switch (e.keyCode) {
			case 40: // ArrowDown
				idx = this._normalizedSelectionIndex();
				max = this._numResults();

				this.setState({
					selectionIndex: idx + 1 >= max ? 0 : idx + 1,
				} as State, this._scrollToVisibleSelection);

				this._temporarilyIgnoreMouseSelection();
				e.preventDefault();
				break;

			case 38: // ArrowUp
				idx = this._normalizedSelectionIndex();
				max = this._numResults();

				this.setState({
					selectionIndex: idx <= 0 ? max - 1 : idx - 1,
				} as State, this._scrollToVisibleSelection);

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

	_scrollToVisibleSelection(): void {
		if (this._selectedItem) {
			(ReactDOM.findDOMNode(this._selectedItem) as HTMLElement).scrollIntoView(false);
		}
	}

	_setSelectedItem(e: any): void {
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
	_onSelection(trackOnly: boolean): void {
		const i = this._normalizedSelectionIndex();
		if (i === -1) {
			return;
		}

		let eventProps: any = {
			globalSearchQuery: this.state.query,
			indexSelected: i,
		};

		let offset = 0;
		if (this.state.matchingResults.Repos) {
			if (i < this.state.matchingResults.Repos.length) {
				const url = `/${this.state.matchingResults.Repos[i].URI}`;
				eventProps.selectedItem = url;
				EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_GLOBAL_SEARCH, AnalyticsConstants.ACTION_CLICK, "GlobalSearchItemSelected", eventProps);
				if (!trackOnly) {
					this._navigateTo(url);
				}
				return;
			}

			offset = this.state.matchingResults.Repos.length;
		}

		const def = this.state.matchingResults.Defs[i - offset];
		let url = urlToDefInfo(def);
		url = url.replace(/GoPackage\/pkg\//, "GoPackage/"); // TEMP HOTFIX

		eventProps.selectedItem = url;
		eventProps.totalResults = this.state.matchingResults.Defs.length;
		if (def.FmtStrings && def.FmtStrings.Kind && def.FmtStrings.Language && def.Repo) {
			eventProps = Object.assign({}, eventProps, { languageSelected: def.FmtStrings.Language, kindSelected: def.FmtStrings.Kind, repoSelected: def.Repo });
		}

		EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_GLOBAL_SEARCH, AnalyticsConstants.ACTION_CLICK, "GlobalSearchItemSelected", eventProps);

		if (!trackOnly) {
			this._navigateTo(url);
		}
	}

	_selectItem(i: number): void {
		this.setState({
			selectionIndex: i,
		} as State);
	}

	// _mouseSelectItem causes i to be selected ONLY IF the user is using the
	// mouse to select. It ignores the case where the user is using the up/down
	// keys to change the selection and the window scrolls, causing the mouse cursor
	// to incidentally hover a different element. We ignore mouse selections except
	// those where the mouse was actually moved.
	_mouseSelectItem(ev: React.MouseEvent<{}>, i: number): void {
		if (this._ignoreMouseSelection) {
			return;
		}
		this._selectItem(i);
	}

	// _temporarilyIgnoreMouseSelection is used to ignore mouse selections. See
	// _mouseSelectItem.
	_temporarilyIgnoreMouseSelection(): void {
		if (!this._debouncedUnignoreMouseSelection) {
			this._debouncedUnignoreMouseSelection = debounce(() => {
				this._ignoreMouseSelection = false;
			}, 200, { leading: false, trailing: true });
		}
		this._debouncedUnignoreMouseSelection();
		this._ignoreMouseSelection = true;
	}

	_results(): JSX.Element[] | null {
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
					ref={selected ? this._setSelectedItem : undefined}
					to={urlToRepo(repo.URI)}
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
			let defURL = urlToDefInfo(def);

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
			const name = qualifiedNameAndType(def, { namequal: "depqualified" });
			list.push(
				<Link className={classNames(styles.block, selected ? styles.result_selected : styles.result, this.state.resultClassName)}
					onMouseOver={(ev) => this._mouseSelectItem(ev, i)}
					ref={selected ? this._setSelectedItem : undefined}
					to={defURL.replace(/GoPackage\/pkg\//, "GoPackage/")}
					key={defURL}
					target="_self"
					onClick={() => this._onSelection(true)}>
					<div className={classNames(styles.cool_gray, styles.flex_container, base.pt3)}>
						<div className={classNames(styles.flex, styles.w100)}>
							<p className={classNames(styles.cool_mid_gray, styles.block_s, base.ma0, base.pl4, base.pr2, base.fr)}>{trimRepo(def.Repo)}</p>
							<code className={classNames(styles.block, styles.f5, base.pb3)}>
								{name}
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
		return <div className={classNames(styles.center, styles.flex, this.state.className)}>
			{this._results()}
			<EventListener target={global.document} event="keydown" callback={this._handleKeyDown} />
		</div>;
	}
}

type Qual = "DepQualified" | "ScopeQualified";

function qualifiedNameAndType(def: any, opts?: any): any {
	if (!def) {
		throw new Error("def is null");
	}
	if (!def.FmtStrings) {
		return "(unknown def)";
	}
	const qual: Qual = opts && opts.nameQual ? opts.nameQual : "ScopeQualified";
	const f = def.FmtStrings;

	let name = f.Name[qual];
	if (f.Name.Unqualified && name) {
		let parts = name.split(f.Name.Unqualified);
		name = [
			parts.slice(0, parts.length - 1).join(f.Name.Unqualified),
			<span key="unqualified" className={opts && opts.unqualifiedNameClass}>{f.Name.Unqualified}</span>,
		];
	}

	return [
		f.DefKeyword,
		f.DefKeyword ? " " : "",
		<span key="name" className={opts && opts.nameClass} style={opts && opts.nameClass ? {} : { fontWeight: "bold" }}>{name}</span>, // give default bold styling if not provided
		f.NameAndTypeSeparator,
		f.Type.ScopeQualified,
	];
}
