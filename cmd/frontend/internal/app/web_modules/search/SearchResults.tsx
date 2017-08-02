import { SearchResult, searchText } from "app/backend";
import { ReferencesGroup } from "app/references/ReferencesWidget";
import { getSearchParamsFromURL } from "app/search";
import { normalFontColor } from "app/util/colors";
import * as React from "react";
import { style } from "typestyle";
import * as URI from "urijs";
import * as activeRepos from "app/util/activeRepos";

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
		let repos = params.repos.replace(/^[,\s]+|[,\s]+$/g, '');
		repos = repos.replace(/\s*,\s*/g, ',');

		// Split the list of repos, and create "active" and "active-and-inactive"
		// booleans + remove them from the list.
		let repoList: string[] = [];
		let addActive, addInactive = false;
		repos.split(',').forEach((repo) => {
			if(repo === "active") {
				addActive = true;
				return;
			}
			if(repo == "active-and-inactive") {
				addActive = true;
				addInactive = true;
				return;
			}
			repoList.push(repo)
		})

		const start = Date.now();
		let search = (repoList) => {
			searchText(q, repoList.map(repo => ({ repo, rev: "" })), params).then(res => {
				const searchDuration = Date.now() - start;
				if (res.results) {
					this.setState({ results: res.results, loading: false, searchDuration });
				}
			});
		};

		// If we need to add active or inactive repositories to the list, do so
		// inside the promise:
		if (addActive || addInactive) {
			activeRepos.get().then((activeRepos) => {
				if (addActive) {
					activeRepos.active.forEach((active) => {
						repoList.push(active);
					})
				}
				if (addInactive) {
					activeRepos.inactive.forEach((inactive) => {
						repoList.push(inactive);
					})
				}
				search(repoList);
			}).catch((error) => {
				// TODO: actually tell the user about the error.
				console.error("failed to get active repos:", error);
				this.setState({loading: false});
			})
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
