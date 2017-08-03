import * as csstips from "csstips";
import * as React from "react";
import * as AddIcon from "react-icons/lib/md/add";
import * as CheckboxFilled from "react-icons/lib/md/check-box";
import * as CheckboxOutline from "react-icons/lib/md/check-box-outline-blank";
import * as SearchIcon from "react-icons/lib/md/search";
import { fetchRepos } from "sourcegraph/backend";
import { Autocomplete } from "sourcegraph/components/Autocomplete";
import { getSearchParamsFromLocalStorage, getSearchPath, handleSearchInput, SearchParams } from "sourcegraph/search";
import { inputBackgroundColor, normalFontColor, primaryBlue, searchFrameBackgroundColor, white } from "sourcegraph/util/colors";
import { style } from "typestyle";

// import * as scrollIntoView from "dom-scroll-into-view";
import * as scrollIntoViewIfNeeded from "scroll-into-view-if-needed";

namespace Styles {
	const padding = "10px";
	const borderRadius = "2px";
	const rowHeight = "32px";

	const input = { backgroundColor: inputBackgroundColor, padding, border: "none", color: white, fontSize: "14px", $nest: { "&::placeholder": { color: normalFontColor }, "&:focus": { outline: "none" } } };

	export const form = style({ backgroundColor: searchFrameBackgroundColor, padding: "16px", borderRadius: "4px" });

	export const searchRow = style(csstips.horizontal);
	export const searchInput = style(input, csstips.flex, { borderRadius, height: rowHeight, marginRight: "15px" });
	export const searchButton = style(csstips.horizontal, csstips.center, csstips.content, { backgroundColor: primaryBlue, height: rowHeight, padding, borderRadius, color: `${white} !important`, textDecoration: "none" });

	export const icon = style({ fontSize: "18px", marginRight: "8px" });

	export const reposSection = style({ marginTop: "16px" });
	export const reposInput = style(input, { marginTop: "8px", borderRadius, minHeight: "64px", maxHeight: "250px", width: "100%", maxWidth: "100%" });
	export const addReposButton = style(csstips.flex, csstips.horizontal, csstips.center, { marginTop: "8px", backgroundColor: inputBackgroundColor, height: rowHeight, padding, cursor: "pointer", borderRadius });

	export const autocomplete = style({ marginTop: "8px", backgroundColor: inputBackgroundColor, cursor: "pointer", borderRadius: "4px", border: "1px solid #2A3A51" });
	export const autocompleteResults = style({ maxHeight: "200px", overflow: "auto" });
	export const addReposInput = style(input, { height: rowHeight, padding, borderRadius: "4px", width: "100%" });
	export const repoSelection = style({ backgroundColor: "#1C2736", color: white, padding: "4px 10px" });
	export const repoSelectionSelected = style({ backgroundColor: "#2A3A51", color: white, padding: "4px 10px" });

	export const filesSection = style({ marginTop: "16px" });
	export const filesInput = style(input, { marginTop: "8px", borderRadius, height: rowHeight, width: "100%" });

	export const filtersSection = style(csstips.horizontal, csstips.center, { marginTop: "16px" });
	export const filter = style(csstips.content, csstips.horizontal, csstips.center, { cursor: "pointer", marginRight: "16px", userSelect: "none" });
}

interface Props {
}

interface State extends SearchParams {
	showAutocomplete: boolean;
}

export interface RepoResult {
	description: string;
	fork: boolean;
	private: boolean;
	pushedAt: string;
	uri: string;
}

export class SearchForm extends React.Component<Props, State> {

	constructor(props: Props) {
		super(props);
		this.state = {
			...getSearchParamsFromLocalStorage(),
			showAutocomplete: false,
		};
	}

	onChange(query: string): void {
		query = query.toLowerCase();
		if (query === "") {
			(this.refs.autocomplete as any).setItems([{ uri: "active" }, { uri: "inactive" }]);
			return;
		}
		fetchRepos(query).then(repos => {
			(this.refs.autocomplete as any).setItems(repos);
		});
	}

	onSelect(item: RepoResult): void {
		const current = this.state.repos.split(/,\s*/);
		let addition = ", " + item.uri;
		for (const uri of current) {
			if (uri === item.uri) {
				addition = "";
				break;
			}
		}
		window.localStorage.setItem("searchRepoScope", this.state.repos + addition);
		this.setState({ showAutocomplete: false, repos: this.state.repos + addition });
	}

	onUpdateRepos(value: string): void {
		window.localStorage.setItem("searchRepoScope", value);
		this.setState({ ...this.state, repos: value });
	}

	render(): JSX.Element | null {
		return <div className={Styles.form}>
			<div className={Styles.searchRow}>
				<input className={Styles.searchInput} autoFocus placeholder="Search..." value={this.state.query} onKeyDown={(e) => handleSearchInput(e, { ...this.state } as any)} onChange={(e) => {
					window.localStorage.setItem("searchQuery", e.target.value);
					this.setState({ ...this.state, query: e.target.value });
				}} />
				<a className={Styles.searchButton} href={getSearchPath(this.state)}>
					<SearchIcon className={Styles.icon} />
					Search code
				</a >
			</div>
			<div className={Styles.reposSection}>
				<div>Repositories</div>
				<textarea className={Styles.reposInput} value={this.state.repos} onChange={(e) => {
					this.onUpdateRepos(e.target.value);
				}} />
				{
					!this.state.showAutocomplete &&
					<div className={Styles.addReposButton} onClick={() => this.setState({ showAutocomplete: true })}>
						<AddIcon className={Styles.icon} />
						<span>Select repositories...</span>
					</div>
				}
				{
					this.state.showAutocomplete &&
					<Autocomplete
						ref="autocomplete"
						ItemView={RepoResult}
						onEscape={() => this.setState({ showAutocomplete: false })}
						className={Styles.autocomplete}
						inputClassName={Styles.addReposInput}
						autocompleteResultsClassName={Styles.autocompleteResults}
						emptyClassName={Styles.repoSelection}
						onChange={(query) => this.onChange(query)}
						onSelect={(item) => this.onSelect(item)}
						onMount={() => setTimeout(() => this.onChange(""), 25)}
						emptyMessage="No results" />
				}
			</div>
			<div className={Styles.filesSection}>
				<div>Files to include</div>
				<input className={Styles.filesInput} value={this.state.files} placeholder="example: *.go" onChange={(e) => {
					window.localStorage.setItem("searchFileScope", e.target.value);
					this.setState({ ...this.state, files: e.target.value });
				}} />
			</div>
			<div className={Styles.filtersSection}>
				{
					[{ key: "matchCase", label: "Match case" }, { key: "matchWord", label: "Match word" }, { key: "matchRegex", label: "Regex" }]
						.map((filter, i) => {
							const clickHandler = () => {
								const newValue: boolean = !this.state[filter.key];
								window.localStorage.setItem("search" + filter.key[0].toUpperCase() + filter.key.substr(1), "" + newValue);
								this.setState({ [filter.key]: !this.state[filter.key] } as any);
							};
							return <div key={i} className={Styles.filter} onClick={clickHandler}>
								{this.state[filter.key] ? <CheckboxFilled className={Styles.icon} /> : <CheckboxOutline className={Styles.icon} />}
								{filter.label}
							</div>;
						})
				}
			</div>
		</div>;
	}
}

export function RepoResult(props: { highlighted: boolean, item: RepoResult }): JSX.Element | null {
	return <div className={props.highlighted ? Styles.repoSelectionSelected : Styles.repoSelection} ref={(el) => {
		if (props.highlighted && el) {
			{/* scrollIntoView(el, document.querySelector("#autocomplete"), { alignWithTop: false, onlyScrollIfNeeded: true }); */ }
			(scrollIntoViewIfNeeded as any).default(el, false);
		}
	}}>
		{props.item.uri}
	</div>;
}
