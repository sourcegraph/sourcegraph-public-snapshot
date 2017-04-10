import gql from "graphql-tag";
import * as hash from "object-hash";
import * as React from "react";

import { QueryBar } from "sourcegraph/search/QueryBar";
import { ResultsView } from "sourcegraph/search/ResultsView";
import { Query } from "sourcegraph/search/types";
import { gqlClient } from "sourcegraph/util/gqlClient";
import { Dispatcher, Disposables } from "sourcegraph/workbench/utils";

interface P {
}

interface S {
	loading: boolean;
	results?: GQL.ISearchResults;
}

export class SearchPage extends React.Component<P, S> {
	dispatcher: Dispatcher<Query> = new Dispatcher<Query>();
	toDispose: Disposables = new Disposables();
	GQLQuery: any = gql`
	query SearchText(
		$pattern: String!,
		$maxResults: Int!,
		$isRegExp: Boolean!,
		$isWordMatch: Boolean!,
		$repositories: [String!]!,
		$isCaseSensitive: Boolean!,
	) {
		root {
    		searchRepos(
				repositories: $repositories,
				query: {
					pattern: $pattern,
					isRegExp: $isRegExp,
					maxResults: $maxResults,
					isWordMatch: $isWordMatch,
					isCaseSensitive: $isCaseSensitive,
			}) {
            	hasNextPage
				results {
					resource
					lineMatches {
						preview
						lineNumber
						offsetAndLengths
					}
				}
          	}
        }
	}`;

	state: S = {
		loading: false,
	};

	/** The hashed value of the query that we currently want to display. */
	currentSearch: string;

	componentDidMount(): void {
		this.toDispose.add(this.dispatcher.subscribe(this.searchTriggered));
	}

	componentWillUnmount(): void {
		this.toDispose.dispose();
	}

	private repositoriesToSearch(): string[] {
		// TODO this should read from the filters.
		if (localStorage.repositoriesToSearch) {
			return JSON.parse(localStorage.repositoriesToSearch);
		}
		return ["github.com/gorilla/mux", "github.com/gorilla/schema"];
	}

	private searchTriggered = async (query: Query) => {
		this.setState({ ...this.state, loading: true });
		const key = hash(query);
		this.currentSearch = key;
		const response = await gqlClient.query<GQL.IQuery>({
			query: this.GQLQuery,
			variables: {
				...query.query,
				repositories: this.repositoriesToSearch(),
				maxResults: 1000,
			},
		});
		if (!response.data.root) {
			return;
		}
		this.searchFinished(key, response.data.root.searchRepos);
	}

	private searchFinished(queryHash: string, response: GQL.ISearchResults): void {
		if (queryHash !== this.currentSearch) {
			return;
		}
		this.setState({
			loading: false,
			results: response,
		});
	}

	render(): JSX.Element {
		return <div>
			<QueryBar dispatcher={this.dispatcher} />
			<ResultsView loading={this.state.loading} results={this.state.results} />
		</div>;
	}
}
