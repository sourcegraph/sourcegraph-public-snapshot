import * as cloneDeep from "lodash/cloneDeep";
import * as debounce from "lodash/debounce";
import * as findIndex from "lodash/findIndex";
import * as throttle from "lodash/throttle";
import * as React from "react";
import { InjectedRouter } from "react-router";

import { EventListener } from "sourcegraph/Component";
import { Input } from "sourcegraph/components/Input";
import { Search as SearchIcon } from "sourcegraph/components/symbols";
import { Spinner as LoadingIcon } from "sourcegraph/components/symbols";
import { colors } from "sourcegraph/components/utils/index";

import { URIUtils } from "sourcegraph/core/uri";

import { urlToBlob, urlToBlobLineCol } from "sourcegraph/blob/routes";
import * as Dispatcher from "sourcegraph/Dispatcher";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import { RepoStore } from "sourcegraph/repo/RepoStore";
import "string_score";

import { Hint, ResultCategories } from "sourcegraph/quickopen/Components";

import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

const modalStyle = {
	position: "fixed",
	top: 0,
	right: 0,
	left: 0,
	maxWidth: 515,
	margin: "0 auto",
	borderRadius: "0 0 3px 3px",
	backgroundColor: colors.coolGray2(),
	padding: 16,
	display: "flex",
	flexDirection: "column",
	zIndex: 4,
	maxHeight: "90vh",
	fontSize: "1rem",
	boxShadow: `0 2px 4px 0 ${colors.black(.1)}`,
};

export interface Result {
	title: string;
	description: string;
	index?: number;  // Index of the matched segment.
	length?: number; // Length of the matched segment.

	// The URL to the result.
	URLPath: string;
};

interface Props {
	dismissModal: (resultSubmitted?: boolean) => void;
	repo: null | {
		URI: string;
		rev: string | null;
	};
	commitID: string | null;
	files: GQL.IFile[];
	languages: string[];
};

interface Results {
	repos: Category;
	symbols: Category;
	files: Category;
};

interface State {
	// The search string in the input box.
	input: string;
	// The row and category that a user has navigated to using a keyboard.
	selected: {
		category: number;
		row: number;
	};
	// Number of results to show in each category.
	limitForCategory: number[];
	// The results of the search.
	results: Results;
	// Whether or not to allow scrolling. Used to prevent jumping when expanding
	// a category.
	allowScroll: boolean;
};

export interface Category {
	Title: string;
	Results: Result[];
	IsLoading: boolean;
}

export interface SearchDelegate {
	dismiss: any;
	select: (category: number, row: number) => void;
	expand: (category: number) => void;
}

// resultsToArray converts from a structured data type into one that can be used
// by the view.
function resultsToArray(results: Results): Category[] {
	const {symbols, files, repos} = results;
	return [files, symbols, repos];
}

// SearchContainer contains the logic that deals with navigation and data
// fetching.
export class Container extends React.Component<Props, State> {

	static contextTypes: any = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: InjectedRouter };
	searchInput: HTMLElement;
	delegate: SearchDelegate;
	listeners: { remove: () => void }[];

	// fetchRepoResults fetches repository search results, augmented
	// with results from a GitHub search API request. Due to the
	// GitHub search API abuse limit, this is debounced (see the
	// initialization in the constructor).
	fetchRepoResultsWithGitHub: (query: string) => void;

	constructor(props: Props) {
		super(props);
		this.keyListener = this.keyListener.bind(this);
		this.bindSearchInput = this.bindSearchInput.bind(this);
		this.updateInput = this.updateInput.bind(this);
		this.updateResults = throttle(this.updateResults.bind(this), 150, { leading: true, trailing: true });
		this.fetchResults = debounce(this.fetchResults, 100, { leading: true, trailing: true });
		this.state = {
			input: "",
			selected: { category: 0, row: 0 },
			limitForCategory: [3, 3, 3],
			results: {
				symbols: { Title: "Definitions", IsLoading: false, Results: [] },
				files: { Title: "Files", IsLoading: false, Results: [] },
				repos: { Title: "Repositories", IsLoading: false, Results: [] },
			},
			allowScroll: true,
		};
		this.delegate = {
			dismiss: props.dismissModal,
			select: this.select.bind(this),
			expand: this.expand.bind(this),
		};
		this.fetchRepoResultsWithGitHub = debounce(((query: string) => {
			Dispatcher.Backends.dispatch(new RepoActions.WantRepos(this.repoListQueryString(query, true)));
		}).bind(this), 1000, { leading: false, trailing: true }); // 2 second debounce
	}

	componentDidMount(): void {
		this.listeners = [
			RepoStore.addListener(this.updateResults),
		];
		setTimeout(() => {
			this.fetchResults();
			this.updateResults();
		});
	}

	componentWillUnmount(): void {
		this.listeners.forEach(s => { s.remove(); });
	}

	componentWillUpdate(_: Props, nextState: State): void {
		if (nextState.input !== this.state.input) {
			nextState.limitForCategory = [3, 3, 3];
			nextState.selected = { category: 0, row: 0 };
		}
	}

	componentDidUpdate(_: Props, prevState: State): void {
		if (this.state.input !== prevState.input) {
			this.fetchResults();
			this.updateResults();
		}
	}

	keyListener(event: KeyboardEvent): void {
		const results = resultsToArray(this.state.results);
		const visibleRowsInCategory = results.map((r, i) => r.Results ? Math.min(r.Results.length, this.state.limitForCategory[i]) : 0);
		let category = this.state.selected.category;
		let row = this.state.selected.row;
		const nextVisibleRow = direction => {
			let next = category;
			for (let i = 0; i < results.length; i++) {
				next += direction;
				if (visibleRowsInCategory[next] > 0) {
					return next;
				}
			}
			return category;
		};
		if (event.keyCode === 38) { // ArrowUp.
			if (row === 0 && category === 0) {
				// noop
			} else if (row <= 0) {
				const newCategory = nextVisibleRow(-1);
				if (newCategory !== category) {
					row = visibleRowsInCategory[newCategory] - 1;
					category = newCategory;
				}
			} else {
				row--;
			}
		} else if (event.keyCode === 40) { // ArrowDown.
			if (row === visibleRowsInCategory[category] - 1 && category === results.length - 1) {
				// noop
			} else if (row >= visibleRowsInCategory[category] - 1) {
				const next = nextVisibleRow(1);
				row = next === category ? row : 0;
				category = next;
			} else {
				row++;
			}
		} else if (event.keyCode === 13) { // Enter.
			this.select(this.state.selected.category, this.state.selected.row);
		} else {
			return;
		}
		let state = Object.assign(this.state, {
			selected: { category: category, row: row },
			allowScroll: true,
		});
		this.setState(state);
		event.preventDefault();
	}

	fetchResults(): void {
		const query = this.state.input;

		// Fetch results without incurring a GitHub search request immediately
		Dispatcher.Backends.dispatch(new RepoActions.WantRepos(this.repoListQueryString(query)));
		// Debounced fetch of results with GitHub search request
		this.fetchRepoResultsWithGitHub(query);

		if (this.props.repo !== null && this.props.commitID) {
			Dispatcher.Backends.dispatch(new RepoActions.WantSymbols(this.props.languages, this.props.repo.URI, this.props.commitID, query));
		}
	}

	repoListQueryString(query: string, withRemote: boolean = false): string {
		if (withRemote) {
			return `Query=${encodeURIComponent(query)}&RemoteSearch=t&PerPage=100`;
		}
		return `Query=${encodeURIComponent(query)}&PerPage=100`;
	}

	updateInput(event: React.FormEvent<HTMLInputElement>): void {
		const input = (event.target as any).value;
		const state = Object.assign({}, this.state, {
			input: input,
		});
		this.setState(state);
	}

	select(c: number, r: number): void {
		const categories = resultsToArray(this.state.results);
		const results = categories[c];
		if (results && results.Results) {
			const result = results.Results[r];
			const resultInfo = {
				category: c,
				rankInCategory: r,
			};
			Object.assign(resultInfo, result);
			const eventProps = {
				repo: this.props.repo,
				quickOpenResult: resultInfo,
				quickOpenQuery: this.state.input,
			};
			AnalyticsConstants.Events.QuickopenItem_Selected.logEvent(eventProps);
			const url = result.URLPath;
			this.props.dismissModal(false);
			this.context.router.push(url);
		}
	}

	bindSearchInput(node: HTMLElement): void { this.searchInput = node; }

	updateResults(): void {
		const query = this.state.input;
		const repo = this.props.repo;
		let {symbols, files, repos} = cloneDeep(this.state.results);

		// Update symbols
		if (repo && this.props.commitID) {
			const updatedSymbols = RepoStore.symbols.list(this.props.languages, repo.URI, this.props.commitID, query);
			if (updatedSymbols.results.length > 0 || !updatedSymbols.loading) {
				const symbolResults: Result[] = [];
				updatedSymbols.results.forEach(sym => {
					let title = sym.name;
					if (sym.containerName) {
						title = `${sym.containerName}.${sym.name}`;
					}
					const kind = symbolKindName(sym.kind);
					const {path} = URIUtils.repoParamsExt(sym.location.uri);
					const desc = `${kind ? kind : ""} in ${path}`;
					let idx = title.toLowerCase().indexOf(query.toLowerCase());
					const line = sym.location.range.start.line;
					const col = sym.location.range.start.character;
					symbolResults.push({
						title: title,
						description: desc,
						index: idx !== -1 ? idx : 0,
						length: idx !== -1 ? query.length : 0,
						URLPath: urlToBlobLineCol(repo.URI, repo.rev, path, line + 1, col + 1),
					});
				});

				symbols.Results = symbolResults;
			}
			symbols.IsLoading = updatedSymbols.loading;

			// Update files
			interface Scorable {
				score: number;
			}
			let fileResults: (Result & Scorable)[] = this.props.files.reduce((acc, file) => rankFile(acc, file.name, query.toLowerCase()), []);
			fileResults.forEach(file => {
				file.URLPath = urlToBlob(repo.URI, repo.rev, file.title);
			});
			fileResults = fileResults.sort((a, b) => b.score - a.score);
			files.IsLoading = false;
			files.Results = fileResults;
		}

		// Update repos. First look up if there are results that include GitHub results.
		let updatedRepos = RepoStore.repos.list(this.repoListQueryString(query, true));
		if (!updatedRepos) {
			// If not, then fall back to results without GitHub requests.
			updatedRepos = RepoStore.repos.list(this.repoListQueryString(query));
		}
		if (updatedRepos) {
			if (updatedRepos.Repos) {
				const repoResults = updatedRepos.Repos.map(({URI}) => ({ title: URI, URLPath: `/${URI}` }));
				repos.IsLoading = false;
				repos.Results = repoResults;
			} else {
				repos.IsLoading = false;
				repos.Results = [];
			}
		} else {
			repos.IsLoading = true;
		}

		// Update selection (don't want to leave the selection on a category
		// that doesn't exist anymore!)
		const results = { symbols: symbols, repos: repos, files: files };
		const resultsArray = resultsToArray(results);
		const sel = this.state.selected;
		if (resultsArray[sel.category].Results.length - 1 < sel.row) {
			sel.row = 0;
			const firstVisibleCategory = Math.max(0, findIndex(resultsArray, (c => c.Results.length > 0)));
			sel.category = firstVisibleCategory;
		}

		this.setState(Object.assign({}, this.state, { results: results }));
	}

	expand(category: number): () => void {
		return () => {
			const state = Object.assign({}, this.state);
			state.limitForCategory[category] += 12;
			state.allowScroll = false;
			this.setState(state);
			if (this.searchInput) {
				this.searchInput.focus();
			}
		};
	}

	loading(): boolean {
		const results = resultsToArray(this.state.results);
		return results.some(category => category.IsLoading);
	}

	render(): JSX.Element {
		const icon = this.loading() ? <LoadingIcon /> : <SearchIcon style={{ fill: colors.coolGray2() }} />;
		const categories = resultsToArray(this.state.results);
		return (
			<div style={modalStyle}>
				<EventListener target={document.body} event={"keydown"} callback={this.keyListener} />
				<div style={{
					backgroundColor: colors.white(),
					borderRadius: 3,
					width: "100%",
					padding: "3px 10px",
					flex: "0 0 auto",
					height: 45,
					display: "flex",
					alignItems: "center",
					flexDirection: "row",
				}}>
					{icon}
					<Input
						id="SearchInput-e2e-test"
						style={{ boxSizing: "border-box", border: "none", flex: "1 0 auto" }}
						placeholder={this.props.repo ? "Search for repositories, files or definitions" : "Search for repositories"}
						value={this.state.input}
						block={true}
						autoFocus={true}
						domRef={this.bindSearchInput}
						onChange={this.updateInput} />
				</div>
				<Hint />
				<ResultCategories categories={categories}
					selection={this.state.selected}
					delegate={this.delegate}
					scrollIntoView={this.state.allowScroll}
					limits={this.state.limitForCategory} />
			</div>
		);
	}
}

// symbolKindName takes in the value of a SymbolInformation.Kind and
// returns the corresponding name for the kind. This is translated
// over from the values of the SymbolKind constants in
// pkg/lsp/service.go.
function symbolKindName(kind: number): string {
	switch (kind) {
		case 1:
			return "file";
		case 2:
			return "module";
		case 3:
			return "namespace";
		case 4:
			return "package";
		case 5:
			return "class";
		case 6:
			return "method";
		case 7:
			return "property";
		case 8:
			return "field";
		case 9:
			return "constructor";
		case 10:
			return "enum";
		case 11:
			return "interface";
		case 12:
			return "function";
		case 13:
			return "variable";
		case 14:
			return "constant";
		case 15:
			return "string";
		case 16:
			return "number";
		case 17:
			return "boolean";
		case 18:
			return "array";
		default:
			return "";
	}
}

// rankFile scores a filepath against a search query, and filters out bad matches.
export const rankFile = (fileResults, file, query) => {
	if (file.length < query.length) {
		// Return early to avoid an expensive query.
		return fileResults;
	}
	if (query === "") {
		// If we don't have a query, no sense scoring anything.
		fileResults.push({ title: file, description: "", index: -1, length: undefined, score: 0 });
		return fileResults;
	}

	// Score against the full path.
	const fuzziness = .8;
	const minimumSimilarity = .6;
	let score = file.score(query, fuzziness);

	// Score against the last element of the path.
	const basePathRegExp = /.*\/(.+)/;
	const base: any = basePathRegExp.exec(file);
	const baseScore = base ? base[1].score(query, fuzziness) : 0;
	score = Math.max(score, baseScore);

	// Substring matches, with a bonus for prefix.
	const index = file.toLowerCase().indexOf(query);
	if (index === 0) {
		const prefixScore = .8;
		score = Math.max(prefixScore, score);
	} else if (index > 0) {
		const substringScore = .6;
		score = Math.max(substringScore, score);
	}

	// Don't update the results if it isn't good enough.
	if (score < minimumSimilarity) {
		return fileResults;
	}

	const l = index > -1 ? query.length : undefined;
	fileResults.push({ title: file, description: "", index: index, length: l, score: score });
	return fileResults;
};
