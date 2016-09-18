import * as React from "react";
import {InjectedRouter} from "react-router";

import {colors} from "sourcegraph/components/jsStyles/colors";
import {urlToBlobLine, urlToBlob} from "sourcegraph/blob/routes";
import {CategorySelector, Hint, ResultCategories, SearchInput, SingleCategoryResults, TabbedResults, Tag} from "sourcegraph/search/modal/SearchComponent";
import {RepoRev} from "sourcegraph/search/modal/SearchModal";
import {RepoStore} from "sourcegraph/repo/RepoStore";
import {TreeStore} from "sourcegraph/tree/TreeStore";
import * as TreeActions from "sourcegraph/tree/TreeActions";
import * as Dispatcher from "sourcegraph/Dispatcher";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import {Search as SearchIcon} from "sourcegraph/components/symbols";
import {input as inputStyle} from "sourcegraph/components/styles/input.css";

const modalStyle = {
	position: "fixed",
	top: 60,
	right: 0,
	left: 0,
	maxWidth: 800,
	margin: "0 auto",
	borderRadius: "0 0 8px 8px",
	backgroundColor: colors.coolGray2(),
	padding: 16,
	display: "flex",
	flexDirection: "column",
	zIndex: 1,
	maxHeight: "90vh",
	fontSize: 15,
};

const CategoryCount = 3;
export const enum Category {
	file,
	definition,
	repository,
}

export const categoryNames = new Map([
	[Category.file, ["file", "files"]],
	[Category.definition, ["definition", "definitions"]],
	[Category.repository, ["repository", "repositories"]],
]);

export interface Result {
	title: string;
	description: string;
	index?: number;  // Index of the matched segment.
	length?: number; // Length of the matched segment.

	// The URL to the result.
	URLPath: string;
};

interface Props {
	dismissModal: () => void;
	start: Category | null;
};

interface State {
	// The search string in the input box.
	input: string;

	results2: Category[];




	
	// The results of the search.
	results: Map<Category, Result[]>;

	// The category that a user wants to limit their search to.
	tag: Category | null;

	// The index of the row that a user has navigated to using arrow keys and
	// may activate by pressing enter. Zero is the search form.
	selected: number;

	// The category tab that a user has navigated to.
	tab: Category | null;

	// Save the search input element so we can focus/blur it.
	searchInput: HTMLElement | null;
};

export interface SearchActions {
	updateInput: (event: React.FormEvent<HTMLInputElement>) => void;
	dismiss: () => void;
	viewCategory: (category: Category) => void;
	bindSearchInput: (node: HTMLElement) => void;
	activateResult: (URLPath: string) => void;
	activateTag: (category: Category) => void;
}

// Find the total number of results in all categories
export function deepLength(categories: Map<Category, Result[]>): number {
	let acc = 0;
	categories.forEach(results => {
		acc = acc + results.length;
	});
	return acc;
}

export interface Category2 {
	Title: string;
	Results: Result[];
}

// SearchContainer contains the logic that deals with navigation and data
// fetching.
export class SearchContainer extends React.Component<Props & RepoRev, State> {

	static contextTypes: any = {
		router: React.PropTypes.object.isRequired,
	};

	context: {
		router: InjectedRouter,
	};

	constructor({start, dismissModal}: Props) {
		super();
		this.navigationKeys = this.navigationKeys.bind(this);
		this.state = {
			input: "",
			results: new Map(),	// TODO(bl): remove
			results2: [],
			tag: start,
			selected2: [0, 0],
			selected: 0, //start === null ? 1 : 0, // TODO(bl): remove
			tab: null,	 // TODO(bl): remove
			searchInput: null,	// TODO(bl): remove
		};
		this.actions = {
			updateInput: this.updateInput.bind(this),
			dismiss: dismissModal,
			viewCategory: this.viewCategory.bind(this), // TODO(bl): remove
			bindSearchInput: this.bindSearchInput.bind(this),
			activateResult: this.activateResult.bind(this),
			activateTag: this.activateTag.bind(this),
		};
	}

	stores(): Store<any>[] {
		return [RepoStore];
	}

	componentWillMount(): void {
		document.body.addEventListener("keydown", this.navigationKeys);
	}

	componentWillUnmount(): void {
		document.body.removeEventListener("keydown", this.navigationKeys);
	}

	componentWillReceiveProps(nextProps: Props): void {
		if (this.props.start === null && nextProps.start !== null) {
			this.setState(Object.assign({}, this.state, {
				tag: nextProps.start,
				selected: 0,
			}));
		}
	}

	componentDidUpdate(_: Props, prevState: State): void {}

	query(): string {
		return this.state.input.toLowerCase();
	}

	updateResults(): void {
		Dispatcher.Backends.dispatch(new RepoActions.WantSymbols(this.props.repo, this.props.commitID, this.query()));
		Dispatcher.Backends.dispatch(new RepoActions.WantRepos(this.query()));
		Dispatcher.Backends.dispatch(new TreeActions.WantFileList(this.props.repo, this.props.commitID));
	}

	/////////////////// bl-cursor
	
	navigationKeys(event: KeyboardEvent): void {
		if (event.key === "ArrowUp") {
			let selected = this.state.selected2.slice();
			selected[1]--;
			let category = selected[0];
			let elements = this.state.results.get(category);
			if (elements) {
				if (selected[1] < 0) {
					if (category == 0) {
						selected[1]++; // don't go down any further if at min
					} else {
						selected[0]--; // go to previous category
						selected[1] = this.state.results.get(category-1).length - 1;
					}
				}
				let state = Object.assign({}, this.state);
				state.selected2 = selected;

				this.setState(state);
			}
		} else if (event.key === "ArrowDown" && this.state.selected < this.visibleResults()) {
			let selected = this.state.selected2.slice();
			selected[1]++;
			let category = selected[0];
			let elements = this.state.results.get(category);
			if (elements) {
				if (selected[1] >= elements.length) {
					if (category == this.state.results.length - 1) {
						selected[1]--; // don't go down any further if at max
					} else {
						selected[0]++; // advance to next category
						selected[1] = 0;
					}
				}
				let state = Object.assign({}, this.state);
				state.selected2 = selected;

				this.setState(state);
			}
		} else if (event.key === "Enter") {
			this.activateResult("FIXME");
		}
	}

	visibleResults(): number {
		if (this.query() === "" && this.state.tag === null) {
			return CategoryCount;
		}
		return deepLength(this.state.results);
	}

	updateInput(event: {target: {value: string}}): void {
		const input = event.target.value;
		const state = Object.assign({}, this.state, {
			input: input,
			selected: 0,
		});
		this.setState(state);
		this.updateResults();
	}

	activateTag(category: Category): void {
		if (this.state.tag === null && this.query() === "") {
			this.setState(Object.assign({}, this.state, {
				tag: category,
				selected: 0,
			}));
			return;
		}
	}

	activateResult(URLPath: string): void {
		this.context.router.push(URLPath);
		this.props.dismissModal();
	}

	viewCategory(category: Category): void {
		const state = Object.assign({}, this.state, {tab: category});
		this.setState(state);
	}

	bindSearchInput(node: HTMLElement): void {
		const state = Object.assign({}, this.state, {searchInput: node});
		this.setState(state);
		if (this.state.selected === 0 && node) {
			node.focus();
		}
	}

	focusSearchBar(): void {
		if (this.state.searchInput) {
			this.state.searchInput.focus();
		}
	}

	blurSearchBar(): void {
		if (this.state.searchInput) {
			this.state.searchInput.blur();
		}
	}

	componentDidMount(): void {
		if (this.state.selected === 0) {
			this.focusSearchBar();
		}
	}

	render(): JSX.Element {
		let results: Map<Category, Result[]> = new Map();
		let query = this.query();

		const symbols = RepoStore.symbols.list(this.props.repo, this.props.commitID, query);
		if (symbols && this.state.results) {
			let symbolResults = [];
			for (let i = 0; i < symbols.length; i++) {
				const path = symbols[i].location.uri;
				const line = symbols[i].location.range.start.line;
				const idx = symbols[i].name.toLowerCase().indexOf(query.toLowerCase());
				symbolResults.push({
					title: symbols[i].name,
					description: "",
					index: idx,
					length: query.length,
					URLPath: urlToBlobLine(this.props.repo, this.props.commitID, path, line+1),
				});
			}

			results.set(Category.definition, symbolResults);
		}

		const repos = RepoStore.repos.list(query);
		if (repos) {
			const repoResults = repos.Repos.map(({URI}) => ({title: URI, URLPath: `/${URI}`}));
			results.set(Category.repository, repoResults);
		}

		const files = TreeStore.fileLists.get(this.props.repo, this.props.commitID);
		if (files) {
			let fileResults = [];
			files.Files.forEach((file, i) => {
				let index = file.toLowerCase().indexOf(query.toLowerCase());
				if (index === -1) { return }
				fileResults.push({ title: file, index: index, length: query.length, URLPath: urlToBlob(this.props.repo, null, file) });
			});
			results.set(Category.file, fileResults);
		}

		const data = {
			input: query,
			results: results,
			tab: this.state.tab,
			tag: this.state.tag,
			selected: this.state.selected,
			selected2: this.state.selected2,
			recentItems: results,
		};


		let content;
		let showHint = true;
		if (data.input === "" && data.tag === null) {
			content = <CategorySelector sel={data.selected} />;
		} else if (data.tag !== null) {
			content = <SingleCategoryResults data={data} category={data.tag} />;
		} else if (data.tab !== null) {
			content = <TabbedResults tab={data.tab} results={data.results} />;
			showHint = false;
		} else {
			content = <ResultCategories resultCategories={data.results} limit={15} selection={data.selected2} />;
		}
		return (
			<div style={modalStyle}>
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
				<SearchIcon style={{fill: colors.coolGray2()}} />
				<Tag tag={data.tag} />
				<input className={inputStyle}
					style={{boxSizing: "border-box", border: "none", flex: "1 0 auto"}}
					placeholder="new http request"
					value={data.input}
					ref={this.actions.bindSearchInput}
					onChange={this.actions.updateInput} />
				<button onClick={this.actions.dismiss} style={{display: "inline"}}>x</button>
				</div>
				{showHint && <Hint tag={data.tag} />}
				{content}
			</div>
		);
	}
}
