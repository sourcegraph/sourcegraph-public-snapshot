import * as React from "react";
import { SearchResult, searchText } from "sourcegraph/backend";
import { ReferencesGroup } from "sourcegraph/references/ReferencesWidget";
import { getSearchParamsFromURL, parseRepoList } from "sourcegraph/search";
import * as activeRepos from "sourcegraph/util/activeRepos";
import { normalFontColor } from "sourcegraph/util/colors";
import { style } from "typestyle";
import * as URI from "urijs";

namespace Styles {
	export const header = style({ padding: "10px 16px", color: normalFontColor, fontSize: "16px" });
}

interface Props { }

interface State {
	results: SearchResult[];
	loading: boolean;
	searchDuration?: number;
}

export class SearchResults extends React.Component<Props, State> {

	constructor(props: Props) {
		super(props);
		const params = getSearchParamsFromURL(window.location.href);
		const q = params.query;
		this.state = {
			results: [],
			loading: true,
		};

		// Clean the comma delimited input (remove whitespace / duplicate commas).
		//
		// See https://stackoverflow.com/a/13306993
		let repos = params.repos.replace(/^[,\s]+|[,\s]+$/g, "");
		repos = repos.replace(/\s*,\s*/g, ",");

		// Split the list of repos, and create "active" and "inactive"
		// booleans + remove them from the list.
		const repoList: string[] = [];
		let addActive = false;
		let addInactive = false;
		for (const repo of parseRepoList(repos)) {
			if (repo === "active") {
				addActive = true;
				continue;
			}
			if (repo === "inactive") {
				addActive = true;
				addInactive = true;
				continue;
			}
			repoList.push(repo);
		}

		const start = Date.now();
		const search = (searchReposList) => {
			searchText(q, searchReposList.map(repo => ({ repo, rev: "" })), params).then(res => {
				const searchDuration = Date.now() - start;
				if (res.results) {
					this.setState({ results: res.results, loading: false, searchDuration });
				}
			});
		};

		// If we need to add active or inactive repositories to the list, do so
		// inside the promise:
		if (addActive || addInactive) {
			activeRepos.get().then((r) => {
				if (addActive) {
					r.active.forEach((active) => {
						repoList.push(active);
					});
				}
				if (addInactive) {
					r.inactive.forEach((inactive) => {
						repoList.push(inactive);
					});
				}
				search(repoList);
			}).catch((error) => {
				// TODO: actually tell the user about the error.
				console.error("failed to get active repos:", error);
				this.setState({ loading: false });
			});
		} else {
			// Don't need to add active or inactive repositories, so perform
			// our search without waiting for the active repo list.
			search(repoList);

			// But also request it, so that it's cached for later.
			activeRepos.get();
		}
	}

	render(): JSX.Element | null {
		if (this.state.loading) {
			return <div className={Styles.header}>
				Working...
			</div>;
		}
		if (!this.state.results || this.state.results.length === 0) {
			return <div className={Styles.header}>
				No results
			</div>;
		}
		return <div>
			<div className={Styles.header}>
				Rendered results in {this.state.searchDuration! / 1000} seconds!
			</div>
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
					};
				});

				return <ReferencesGroup uri={parsed.hostname + parsed.path} path={parsed.fragment} key={i} refs={refs} isLocal={false} />;
			})}
		</div>;
	}
}
