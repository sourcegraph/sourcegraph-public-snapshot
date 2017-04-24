import gql from "graphql-tag";
import * as hash from "object-hash";
import * as React from "react";
import { IModeService } from "vs/editor/common/services/modeService";

import { context, isOnPremInstance } from "sourcegraph/app/context";
import { colors } from "sourcegraph/components/utils";
import { ComponentWithRouter } from "sourcegraph/core/ComponentWithRouter";
import { QueryBar } from "sourcegraph/search/QueryBar";
import { ResultsView } from "sourcegraph/search/ResultsView";
import "sourcegraph/search/SearchPage.css";
import { Query } from "sourcegraph/search/types";
import { gqlClient } from "sourcegraph/util/gqlClient";
import { Services } from "sourcegraph/workbench/services";
import { Dispatcher, Disposables } from "sourcegraph/workbench/utils";

interface P {
}

interface S {
	loading: boolean;
	results?: GQL.ISearchResults;
}

const contentSx = {
	margin: "0 auto",
	maxWidth: "90%",
	width: 800,
};

const pageSx = {
	backgroundColor: colors.blueGrayL3(),
	height: "100%",
	overflow: "overlay",
};

export async function getRepositories(): Promise<GQL.IRepository[]> {
	const query = gql`
			query {
				root {
					repositories {
						uri
						language
					}
				}
			}
		`;
	const response = await gqlClient.query<GQL.IQuery>({
		query: query,
		variables: {
			...query.query,
			maxResults: 500,
		},
	});
	if (!response.data.root) {
		return [];
	}
	return response.data.root.repositories;
}

export class SearchPage extends ComponentWithRouter<P, S> {
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
	initialQuery: string = this.context.router.location.query["q"] || "";
	state: S = {
		loading: false,
	};

	/** The hashed value of the query that we currently want to display. */
	currentSearch: string;

	constructor(props: P, context: any) {
		super(props, context);
		// make sure query is run once on init
		getRepositories();
	}

	async componentDidMount(): Promise<void> {
		this.toDispose.add(this.dispatcher.subscribe(this.searchTriggered));

		// TODO(nicot) this preloads some common languages, to improve perceived search speed.
		const modeService = Services.get(IModeService) as IModeService;
		["go", "java", "javascript", "typescript", "python"].map(mode => modeService.getOrCreateMode(mode));
		if (!this.initialQuery) { return; }
		setTimeout(() => {
			this.doSearch({
				query: {
					pattern: this.initialQuery,
					isCaseSensitive: false,
					isMultiline: false,
					isRegExp: false,
					isWordMatch: false,
				},
			});
		}, 500);
	}

	componentWillUnmount(): void {
		this.toDispose.dispose();
	}

	private async repositoriesToSearch(): Promise<string[]> {
		// TODO this should read from the filters.
		if (localStorage.repositoriesToSearch) {
			return JSON.parse(localStorage.repositoriesToSearch);
		}
		// if on-prem, and not overridden by localstorage, use set of repos visibile to user on dashboard page
		if (isOnPremInstance(context.authEnabled)) {
			const regexFilter = context.repoHomeRegexFilter ? RegExp(context.repoHomeRegexFilter) : null;
			return getRepositories().then(repos => {
				return repos.map(repo => repo.uri).filter(repo => regexFilter ? regexFilter.test(repo) : true);
			});
		}
		return ["github.com/gorilla/mux", "github.com/gorilla/schema"];
	}

	private async doSearch(query: Query): Promise<void> {
		this.setState({ ...this.state, loading: true });
		const key = hash(query);
		this.currentSearch = key;
		if (query.query.pattern === "") {
			return this.searchFinished(key, undefined);
		}
		const reposToSearch = await this.repositoriesToSearch();
		const response = await gqlClient.query<GQL.IQuery>({
			query: this.GQLQuery,
			variables: {
				...query.query,
				repositories: reposToSearch,
				maxResults: 500,
			},
		});
		if (!response.data.root) {
			return;
		}
		this.searchFinished(key, response.data.root.searchRepos);
	}

	private searchTriggered = (query: Query) => {
		const loc = new URL(location.toString());
		loc.search = query.query.pattern;
		history.pushState({}, "Search", loc.toString());
		this.doSearch(query);
	}

	private searchFinished(queryHash: string, response?: GQL.ISearchResults): void {
		if (queryHash !== this.currentSearch) {
			return;
		}
		this.setState({
			loading: false,
			results: response,
		});
	}

	render(): JSX.Element {
		return <div className="search-page" style={pageSx}>
			<div style={contentSx}>
				<QueryBar initialQuery={this.initialQuery} dispatcher={this.dispatcher} />
				<ResultsView loading={this.state.loading} results={this.state.results} />
			</div>
		</div>;
	}
}
