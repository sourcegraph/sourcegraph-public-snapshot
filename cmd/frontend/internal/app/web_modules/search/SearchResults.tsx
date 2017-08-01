import { searchText, SearchResult } from "app/backend";
import { getSearchParamsFromURL } from "app/search";
import { ReferencesGroup } from "app/references/ReferencesWidget";
// import { inputBackgroundColor, normalFontColor, primaryBlue, searchFrameBackgroundColor, white } from "app/util/colors";
// import * as csstips from "csstips";
import * as React from "react";
// import { style } from "typestyle";
import * as URI from "urijs";

// import * as scrollIntoView from "dom-scroll-into-view";


// namespace Styles {
//     const padding = "10px";
//     const borderRadius = "2px";
//     const rowHeight = "32px";

//     const input = { backgroundColor: inputBackgroundColor, padding, border: "none", color: white, fontSize: "14px", $nest: { "&::placeholder": { color: normalFontColor }, "&:focus": { outline: "none" } } };

//     export const form = style({ backgroundColor: searchFrameBackgroundColor, padding: "16px", borderRadius: "4px", })

//     export const searchRow = style(csstips.horizontal)
//     export const searchInput = style(input, csstips.flex, { borderRadius, height: rowHeight, marginRight: "15px" });
//     export const searchButton = style(csstips.horizontal, csstips.center, csstips.content, { backgroundColor: primaryBlue, height: rowHeight, padding, borderRadius, color: `${white} !important`, textDecoration: "none" })

//     export const icon = style({ fontSize: "18px", marginRight: "10px" })

//     export const reposSection = style({ marginTop: "16px" })
//     export const reposInput = style(input, { marginTop: "8px", borderRadius, minHeight: "64px", maxHeight: "250px", width: "100%", maxWidth: "100%" })
//     export const addReposButton = style(csstips.flex, csstips.horizontal, csstips.center, { marginTop: "8px", backgroundColor: inputBackgroundColor, height: rowHeight, padding, cursor: "pointer", borderRadius });

//     export const autocomplete = style({ marginTop: "8px", backgroundColor: inputBackgroundColor, cursor: "pointer", borderRadius: "4px", border: "1px solid #2A3A51" });
//     export const autocompleteResults = style({ maxHeight: "200px", overflow: "auto" });
//     export const addReposInput = style(input, { height: rowHeight, padding, borderRadius: "4px", width: "100%" });
//     export const repoSelection = style({ backgroundColor: "#1C2736", color: white, padding })
//     export const repoSelectionSelected = style({ backgroundColor: "#2A3A51", color: white, padding })

//     export const filesSection = style({ marginTop: "16px" });
//     export const filesInput = style(input, { marginTop: "8px", borderRadius, height: rowHeight, width: "100%" });
// }

interface Props {
}

interface State {
	results: SearchResult[];
}

export class SearchResults extends React.Component<Props, State> {

	constructor(props: Props) {
		super(props);
		const params = getSearchParamsFromURL(window.location.href);
		const repos = params.repos;
		const q = params.query;
		this.state = {
			results: [],
			// query: query["q"] || "",
			// repos: query["repos"] || "",
			// files: query["files"] || "",
			// showAutocomplete: false,
		}

		let split = repos.split(/,\s */);
		searchText(q, split.filter(repo => !repo.startsWith("active")).map(repo => ({ repo, rev: "" })), params).then(res => {
			if (res.results) {
				this.setState({ results: res.results })
			}
		});
	}

	render(): JSX.Element | null {
		if (this.state.results.length === 0) {
			return null;
		}
		return <div>
			{this.state.results.map((result, i) => {
				const parsed = URI.parse(result.resource);
				const refs = result.lineMatches.map(match => {
					return {
						range: {
							start: {
								character: match.offsetAndLengths[0][0],
								line: match.lineNumber,
							},
							end: {
								character: match.offsetAndLengths[0][0] + match.offsetAndLengths[0][1],
								line: match.lineNumber,
							},
						},
						uri: result.resource,
						repoURI: parsed.hostname + parsed.path,
					}
				});

				return <ReferencesGroup uri={parsed.hostname + parsed.path} path={parsed.fragment} key={i} refs={refs} isLocal={false} />;
			})}
		</div>;
	}
}
