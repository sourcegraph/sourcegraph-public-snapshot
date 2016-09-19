import * as React from "react";
import {InjectedRouter} from "react-router";

import {urlToBlob, urlToBlobLine} from "sourcegraph/blob/routes";

import {Input} from "sourcegraph/components/Input";
import {colors} from "sourcegraph/components/jsStyles/colors";
import {Search as SearchIcon} from "sourcegraph/components/symbols";
import {Container} from "sourcegraph/Container";
import {RepoStore} from "sourcegraph/repo/RepoStore";
import {Store} from "sourcegraph/Store";
import {TreeStore} from "sourcegraph/tree/TreeStore";

import {ResultCategories} from "sourcegraph/search/modal/SearchComponent";
import {RepoRev} from "sourcegraph/search/modal/SearchModal";

import * as Dispatcher from "sourcegraph/Dispatcher";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import * as TreeActions from "sourcegraph/tree/TreeActions";

const modalStyle = {
	position: "fixed",
	top: 60,
	right: 0,
	left: 0,
	maxWidth: 515,
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

	// The [category, row] of the currently selected item
	selected: number[];
};

export interface Category {
	Title: string;
	Results?: Result[];
	IsLoading: boolean;
}

export interface SearchDelegate {
	dismiss: any;
	select: (category: number, row: number) => void;
}

// SearchContainer contains the logic that deals with navigation and data
// fetching.
export class SearchContainer extends Container<Props & RepoRev, State> {

	static contextTypes: any = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: InjectedRouter };
	searchInput: HTMLElement;
	delegate: SearchDelegate;

	constructor(props: Props & RepoRev) {
		super(props);
		this.keyListener = this.keyListener.bind(this);
		this.bindSearchInput = this.bindSearchInput.bind(this);
		this.updateInput = this.updateInput.bind(this);
		this.state = {
			input: "",
			results: [],
			selected: [0, 0],
		};
		this.delegate = {
			dismiss: props.dismissModal,
			select: this.select.bind(this),
		};
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

	query(): string {
		return this.state.input.toLowerCase();
	}

	keyListener(event: KeyboardEvent): void {
		let results = this.results();
		let categorySizes = results.map((r) => r.Results ? r.Results.length : 0);
		if (event.key === "ArrowUp") {
			let selected = this.state.selected.slice();
			selected[1]--;
			let c = selected[0];
			if (selected[1] < 0) {
				if (c === 0) {
					selected[1]++; // don't go down any further if at min
				} else {
					selected[0]--; // go to previous category
					selected[1] = categorySizes[c - 1] - 1;
				}
			}
			let state = Object.assign({}, this.state);
			state.selected = selected;
			this.setState(state);
		} else if (event.key === "ArrowDown") {
			let selected = this.state.selected.slice();
			selected[1]++;
			let c = selected[0];
			let cN = categorySizes[c];
			if (selected[1] >= cN) {
				if (c === categorySizes.length - 1) {
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
		} else {
			return;
		}
		event.preventDefault();
	}

	fetchResults(query: string): void {
		Dispatcher.Backends.dispatch(new RepoActions.WantSymbols(this.props.repo, this.props.commitID, query));
		Dispatcher.Backends.dispatch(new RepoActions.WantRepos(this.repoListQueryString(query)));
		Dispatcher.Backends.dispatch(new TreeActions.WantFileList(this.props.repo, this.props.commitID));
	}

	repoListQueryString(query: string): string {
		return `Query=${encodeURIComponent(query)}&Type=public&LocalOnly=true`;
	}

	updateInput(event: React.FormEvent<HTMLInputElement>): void {
		const input = (event.target as any).value;
		const state = Object.assign({}, this.state, {
			input: input,
		});
		this.setState(state);
		this.fetchResults(input.toLowerCase());
	}

	select(c: number, r: number): void {
		let categories = this.results();
		let results = categories[c].Results;
		if (results) {
			let url = results[r].URLPath;
			this.props.dismissModal();
			this.context.router.push(url);
		}
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
			let symbolResults: Result[] = [];
			for (let i = 0; i < symbols.length; i++) {
				let title = symbols[i].name;
				let kind = symbolKindName(symbols[i].kind);
				if (kind) {
					title = `${kind} ${title}`;
				}
				const path = symbols[i].location.uri;
				const line = symbols[i].location.range.start.line;
				const idx = title.toLowerCase().indexOf(query.toLowerCase());
				symbolResults.push({
					title: title,
					description: symbols[i].location.uri,
					index: idx,
					length: query.length,
					URLPath: urlToBlobLine(this.props.repo, this.props.commitID, path, line + 1),
				});
			}

			symbolResults = symbolResults.slice(0, 3);
			results.push({ Title: "Definitions", IsLoading: false, Results: symbolResults });
		} else {
			results.push({ Title: "Definitions", IsLoading: true });
		}

		const repos = RepoStore.repos.list(this.repoListQueryString(query));
		if (repos) {
			if (repos.Repos) {
				let repoResults = repos.Repos.map(({URI}) => ({title: URI, URLPath: `/${URI}`}));
				repoResults = repoResults.slice(0, 3);
				results.push({ Title: "Repositories", IsLoading: false, Results: repoResults });
			} else {
				results.push({ Title: "Repositories", IsLoading: false, Results: [] });
			}
		} else {
			results.push({ Title: "Repositories", IsLoading: true });
		}

		const files = TreeStore.fileLists.get(this.props.repo, this.props.commitID);
		if (files) {
			let fileResults: Result[] = [];
			files.Files.forEach((file, i) => {
				let index = file.toLowerCase().indexOf(query.toLowerCase());
				if (index === -1) {
					return;
				}
				fileResults.push({ title: file, description: "", index: index, length: query.length, URLPath: urlToBlob(this.props.repo, null, file) });
			});
			fileResults = fileResults.slice(0, 3);
			results.push({ Title: "Files", IsLoading: false, Results: fileResults });
		} else {
			results.push({ Title: "Files", IsLoading: true });
		}

		return results;
	}

	render(): JSX.Element {
		let categories = this.results();
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
				<Input
					style={{boxSizing: "border-box", border: "none", flex: "1 0 auto"}}
					placeholder="jump to def, file, or repository"
					value={this.state.input}
					domRef={this.bindSearchInput}
					onChange={this.updateInput} />
				</div>
				<ResultCategories categories={categories} selection={this.state.selected} delegate={this.delegate} />
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
		return "func";
	case 13:
		return "var";
	case 14:
		return "const";
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
