import * as React from "react";
import {InjectedRouter} from "react-router";

import {Container} from "sourcegraph/Container";
import {colors} from "sourcegraph/components/jsStyles/colors";
import {urlToBlobLine, urlToBlob} from "sourcegraph/blob/routes";
import {CategorySelector, ResultCategories, SearchInput, SingleCategoryResults, TabbedResults, Tag} from "sourcegraph/search/modal/SearchComponent";
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
};

interface State {
	// The search string in the input box.
	input: string;

	// The results of the search
	results: any;

	// The index of the row that a user has navigated to using arrow keys and
	// may activate by pressing enter. Zero is the search form.
	selected: number;

	// Save the search input element so we can focus/blur it.
	searchInput: HTMLElement | null;
};

export interface Category {
	Title: string;
	Results: Result[];
}

interface SearchDelegate {
	dismiss: any;
	select: (category: number, row: number) => void;
}

// SearchContainer contains the logic that deals with navigation and data
// fetching.
export class SearchContainer extends Container<Props & RepoRev, State> {

	static contextTypes: any = {
		router: React.PropTypes.object.isRequired,
	};

	context: {
		router: InjectedRouter,
	};

	constructor({start, dismissModal}: Props) {
		super();
		this.keyListener = this.keyListener.bind(this);
		this.bindSearchInput = this.bindSearchInput.bind(this);
		this.updateInput = this.updateInput.bind(this);
		this.state = {
			input: "",
			results: [],
			selected: [0, 0],
		};
		this.delegate = {
			dismiss: dismissModal,
			select: this.select.bind(this),
		}
	}

	stores(): Store<any>[] {
		return [RepoStore];
	}

	reconcileState(state: State, props: Props): void {
		state.results = this.results();
	}

	componentWillMount(): void {
		super.componentWillMount();
		document.body.addEventListener("keydown", this.keyListener);
	}

	componentDidMount(): void {
		super.componentDidMount();
		this.fetchResults("");
		this.focusSearchBar();
	}

	componentWillUnmount(): void {
		super.componentWillUnmount();
		document.body.removeEventListener("keydown", this.keyListener);
	}

	onStateTransition(prevState: State, nextState: State): void {}

	query(): string {
		return this.state.input.toLowerCase();
	}

	keyListener(event: KeyboardEvent): void {
		let results = this.results();
		let categorySizes = results.map((r) => r.Results.length);
		if (event.key === "ArrowUp") {
			let selected = this.state.selected.slice();
			selected[1]--;
			let c = selected[0];
			if (selected[1] < 0) {
				if (c == 0) {
					selected[1]++; // don't go down any further if at min
				} else {
					selected[0]--; // go to previous category
					selected[1] = categorySizes[c-1] - 1;
				}
			}
			let state = Object.assign({}, this.state);
			state.selected = selected;
			this.setState(state);
		} else if (event.key === "ArrowDown") {
			let selected = this.state.selected.slice();
			selected[1]++;
			let c = selected[0];
			let c_n = categorySizes[c];
			if (selected[1] >= c_n) {
				if (c == categorySizes.length - 1) {
					selected[1]--; // don't go down any further if at max
				} else {
					selected[0]++; // advance to next category
					selected[1] = 0;
				}
			}
			let state = Object.assign({}, this.state);
			state.selected = selected;
			this.setState(state);
		} else if (event.key === "Enter") {
			this.select(this.state.selected[0], this.state.selected[1]);
		}
	}

	fetchResults(query: string): void {
		Dispatcher.Backends.dispatch(new RepoActions.WantSymbols(this.props.repo, this.props.commitID, query));
		Dispatcher.Backends.dispatch(new RepoActions.WantRepos(query));
		Dispatcher.Backends.dispatch(new TreeActions.WantFileList(this.props.repo, this.props.commitID));
	}

	updateInput(event: {target: {value: string}}): void {
		const input = event.target.value;
		const state = Object.assign({}, this.state, {
			input: input,
		});
		this.setState(state);
		this.fetchResults(input.toLowerCase());
	}

	select(c: number, r: number): void {
		let categories = this.results();
		let url = categories[c].Results[r].URLPath;
		this.props.dismissModal();
		this.context.router.push(url);
	}

	bindSearchInput(node: HTMLElement): void { this.searchInput = node; }

	focusSearchBar(): void {
		if (this.searchInput) {
			this.searchInput.focus();
		}
	}

	blurSearchBar(): void {
		if (this.searchInput) {
			this.searchInput.blur();
		}
	}

	results(): Category[] {
		let query = this.query();
		let results: Category[] = [];

		const symbols = RepoStore.symbols.list(this.props.repo, this.props.commitID, query);
		if (symbols) {
			let symbolResults = [];
			for (let i = 0; i < symbols.length; i++) {
				const path = symbols[i].location.uri;
				const line = symbols[i].location.range.start.line;
				const idx = symbols[i].name.toLowerCase().indexOf(query.toLowerCase());
				symbolResults.push({
					title: symbols[i].name,
					description: symbols[i].location.uri,
					index: idx,
					length: query.length,
					URLPath: urlToBlobLine(this.props.repo, this.props.commitID, path, line+1),
				});
			}
			symbolResults = symbolResults.slice(0, 3);
			results.push({ Title: "Definitions", Results: symbolResults });
		}

		const repos = RepoStore.repos.list(query);
		if (repos) {
			let repoResults = repos.Repos.map(({URI}) => ({title: URI, URLPath: `/${URI}`}));
			repoResults = repoResults.slice(0, 3);
			results.push({ Title: "Repositories", Results: repoResults });
		}

		const files = TreeStore.fileLists.get(this.props.repo, this.props.commitID);
		if (files) {
			let fileResults = [];
			files.Files.forEach((file, i) => {
				let index = file.toLowerCase().indexOf(query.toLowerCase());
				if (index === -1) { return }
				fileResults.push({ title: file, index: index, length: query.length, URLPath: urlToBlob(this.props.repo, null, file) });
			});
			fileResults = fileResults.slice(0, 3);
			results.push({ Title: "Files", Results: fileResults });
		}

		return results;
	}

	render(): JSX.Element {
		let categories = this.results();
		let query = this.query();
		let content = <ResultCategories categories={categories} limit={15} selection={this.state.selected} delegate={this.delegate} />;
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
				<input className={inputStyle}
					style={{boxSizing: "border-box", border: "none", flex: "1 0 auto"}}
					placeholder="new http request"
					value={this.state.input}
					ref={this.bindSearchInput}
					onChange={this.updateInput} />
				</div>
				{content}
			</div>
		);
	}
}
