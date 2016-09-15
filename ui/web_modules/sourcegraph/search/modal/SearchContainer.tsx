import * as React from "react";

import {SearchComponent} from "sourcegraph/search/modal/SearchComponent";
import {RepoRev} from "sourcegraph/search/modal/SearchModal";
import {updateCategory} from "sourcegraph/search/modal/UpdateResults";

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
	index: number;  // Index of the matched segment. Optional.
	length: number; // Length of the matched segment. Optional.
};

interface Props {
	dismissModal: () => void;
	start: Category | null;
};

interface State {
	// The search string in the input box.
	input: string;

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
	updateInput: (event: KeyboardEvent) => void;
	dismiss: () => void;
	viewCategory: (category: Category) => void;
	bindSearchInput: (node: HTMLElement) => void;
}

// Find the total number of results in all categories
export function deepLength(categories: Map<Category, Result[]>): number {
	let acc = 0;
	categories.forEach(results => {
		acc = acc + results.length;
	});
	return acc;
}

// SearchContainer contains the logic that deals with navigation and data
// fetching.
export class SearchContainer extends React.Component<Props & RepoRev, State> {

	actions: SearchActions;

	constructor({start, dismissModal}: Props) {
		super();
		this.navigationKeys = this.navigationKeys.bind(this);
		this.state = {
			input: "",
			results: new Map(),
			tag: start,
			selected: start === null ? 1 : 0,
			tab: null,
			searchInput: null,
		};
		setTimeout(() => this.fakeResults(), 0);
		this.actions = {
			updateInput: this.updateInput.bind(this),
			dismiss: dismissModal,
			viewCategory: this.viewCategory.bind(this),
			bindSearchInput: this.bindSearchInput.bind(this),
		};
	}

	componentWillMount(): void {
		document.body.addEventListener("keydown", this.navigationKeys);
	}

	componentWillUnmount(): void {
		document.body.removeEventListener("keydown", this.navigationKeys);
	}

	componentDidUpdate(_: Props, prevState: State): void {
		if (this.props.start !== null && this.state.tag  === null) {
			const state = Object.assign(this.state, {
				tag: this.props.start,
				selected: 0,
			});
			this.focusSearchBar();
			this.setState(state);
		}

		if (prevState.selected !== 0 && this.state.selected === 0) {
			this.focusSearchBar();
		}
		if (prevState.selected === 0 && this.state.selected > 0) {
			this.blurSearchBar();
		}
		if (this.state.input !== prevState.input) {
			this.updateResults();
		}
	}

	updateResults(): void {
		for (let i: Category = 0; i < CategoryCount; i++) {
			updateCategory(i, this.props.repo, this.props.commitID, this.state.input, resultList => {
				const results = this.state.results;
				results.set(i, resultList);
				this.setState(Object.assign({}, this.state, {results: results}));
			});
		}
	}

	navigationKeys(event: KeyboardEvent): void {
		if (event.key === "ArrowUp" && this.state.selected > 0) {
			const state = Object.assign({}, this.state, {selected: this.state.selected - 1});
			this.setState(state);
		} else if (event.key === "ArrowDown" && this.state.selected < this.visibleResults()) {
			const state = Object.assign({}, this.state, {selected: this.state.selected + 1});
			this.setState(state);
		} else if (event.key === "Enter") {
			this.activateResult(this.state.selected);
		}
	}

	visibleResults(): number {
		if (this.state.input === "" && this.state.tag === null) {
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
	}

	fakeResults(): void {
		let results = new Array();
		for (let i = 0; i < 20; i++ ) {
			results.push({title: "webpack config", description: "ui/utils some information there hello, only one line though please, thanks"});
		}
		let m = this.state.results;
		m.set(Category.file, results);
		m.set(Category.definition, results);
		m.set(Category.repository, results);
		const state = Object.assign({}, this.state, {results: m});
		this.setState(state);
	}

	activateResult(selected: number): void {
		if (this.state.tag === null && this.state.input === "") {
			this.setState(Object.assign({}, this.state, {
				tag: selected,
				selected: 0,
			}));
			return;
		}
		const URL = "the location of this selection";
		console.error(`navigating to ${URL}`);
		// router.push(URL)
	}

	viewCategory(category: Category): void {
		const state = Object.assign({}, this.state, {tab: category});
		this.setState(state);
	}

	bindSearchInput(node: HTMLElement): void {
		const state = Object.assign({}, this.state, {searchInput: node});
		this.setState(state);
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
		const data = {
			input: this.state.input,
			results: this.state.results,
			tab: this.state.tab,
			tag: this.state.tag,
			selected: this.state.selected,
			recentItems: this.state.results,
		};

		return <SearchComponent
			data={data}
			actions={this.actions} />;
	}
}
