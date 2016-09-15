import * as React from "react";
import {InjectedRouter} from "react-router";

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
	// updateInput: number;
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

// NOTE: this is global. Bad idea?
export let actions: SearchActions = {
	updateInput: (event: React.FormEvent<HTMLInputElement>) => null,
	dismiss: () => null,
	viewCategory: (category: Category) => null,
	bindSearchInput: (node: HTMLElement) => null,
	activateResult: (URLPath: string) => null,
	activateTag: (category: Category) => null,
};

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
			results: new Map(),
			tag: start,
			selected: 0, //start === null ? 1 : 0,
			tab: null,
			searchInput: null,
		};
		actions = {
			updateInput: this.updateInput.bind(this),
			dismiss: dismissModal,
			viewCategory: this.viewCategory.bind(this),
			bindSearchInput: this.bindSearchInput.bind(this),
			activateResult: this.activateResult.bind(this),
			activateTag: this.activateTag.bind(this),
		};
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

	componentDidUpdate(_: Props, prevState: State): void {
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
			this.activateResult("FIXME");
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

	activateTag(category: Category): void {
		if (this.state.tag === null && this.state.input === "") {
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
		const data = {
			input: this.state.input,
			results: this.state.results,
			tab: this.state.tab,
			tag: this.state.tag,
			selected: this.state.selected,
			recentItems: this.state.results,
		};

		return <SearchComponent data={data} />;
	}
}
