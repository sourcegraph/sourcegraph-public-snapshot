import { SearchResult, searchText } from "app/backend";
import { ReferencesGroup } from "app/references/ReferencesWidget";
import { getSearchParamsFromURL } from "app/search";
import { normalFontColor } from "app/util/colors";
import * as React from "react";
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
		const repos = params.repos;
		const q = params.query;
		this.state = {
			results: [],
			loading: true,
		};

		const split = repos.split(/,\s */);
		const start = Date.now();
		searchText(q, split.map(repo => ({ repo, rev: "" })), params).then(res => {
			const searchDuration = Date.now() - start;
			if (res.results) {
				this.setState({ results: res.results, loading: false, searchDuration });
			}
		});
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
