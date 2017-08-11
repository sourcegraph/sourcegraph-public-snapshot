import * as csstips from "csstips";
import * as React from "react";
import { SearchResult, searchText } from "sourcegraph/backend";
import { ReferencesGroup } from "sourcegraph/references/ReferencesWidget";
import { getSearchParamsFromURL, parseRepoList } from "sourcegraph/search";
import * as activeRepos from "sourcegraph/util/activeRepos";
import { normalFontColor, white } from "sourcegraph/util/colors";
import { style } from "typestyle";
import * as URI from "urijs";

namespace Styles {
	export const header = style(csstips.horizontal, csstips.center, { padding: "16px 0px", color: normalFontColor, fontSize: "16px" });
	export const badge = style({ backgroundColor: "#2A3A51 !important", borderRadius: "20px", color: white, marginRight: "8px", fontSize: "11px", padding: "3px 6px", fontFamily: "system" });
	export const label = style({ color: normalFontColor, marginRight: "16px", fontSize: "12px" });
}

interface Props { }

interface State {
	results: SearchResult[];
	loading: boolean;
	searchDuration?: number;
}

function numberWithCommas(x: any): string {
	return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
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
		let totalMatches = 0;
		let totalResults = 0;
		let totalFiles = 0;
		let totalRepos = 0;
		const seenRepos = new Set<string>();
		this.state.results.forEach(result => {
			const parsed = URI.parse(result.resource);
			if (!seenRepos.has(parsed.hostname + parsed.path)) {
				seenRepos.add(parsed.hostname + parsed.path);
				totalRepos += 1;
			}
			totalFiles += 1;
			totalResults += result.lineMatches.length;
		});
		return <div>
			<div className={Styles.header}>
				<div className={Styles.badge}>{numberWithCommas(totalResults)}</div>
				<div className={Styles.label}>results in</div>
				<div className={Styles.badge}>{numberWithCommas(totalFiles)}</div>
				<div className={Styles.label}>files in</div>
				<div className={Styles.badge}>{numberWithCommas(totalRepos)}</div>
				<div className={Styles.label}>repos</div>
			</div>
			{this.state.results.map((result, i) => {
				totalMatches += result.lineMatches.length;
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

				return <ReferencesGroup hidden={totalMatches > 500} uri={parsed.hostname + parsed.path} path={parsed.fragment} key={i} refs={refs} isLocal={false} />;
			})}
		</div>;
	}
}
