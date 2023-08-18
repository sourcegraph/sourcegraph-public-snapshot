import type { ApolloClient } from '@apollo/client'
import gql from 'tagged-template-noop'

import { CodeHostIcon, formatRepositoryStarCount, SearchResultStar } from '@sourcegraph/branded'
import { getDocumentNode } from '@sourcegraph/http-client'

import { getWebGraphQLClient } from '../../backend/graphql'
import type { SearchValue } from '../../fuzzyFinder/SearchValue'
import type { FuzzyFinderRepoResult, FuzzyFinderRepoVariables } from '../../graphql-operations'
import type { UserHistory } from '../useUserHistory'

import { FuzzyStorageCache, type PersistableQueryResult } from './FuzzyLocalCache'
import { FuzzyQuery } from './FuzzyQuery'

export const FUZZY_REPOS_QUERY = gql`
    query FuzzyFinderRepo($query: String!) {
        search(patternType: regexp, query: $query) {
            results {
                repositories {
                    name
                    stars
                }
            }
        }
    }
`

export class FuzzyRepos extends FuzzyQuery {
    constructor(
        private readonly client: ApolloClient<object> | undefined,
        onNamesChanged: () => void,
        private userHistory: UserHistory
    ) {
        super(
            onNamesChanged,
            new FuzzyStorageCache(window.localStorage, 'fuzzy-finder.repository-names', values =>
                this.staleResults(values)
            )
        )
    }

    /* override */ protected rawQuery(query: string): string {
        return `type:repo count:100 ${query.replace('/', '.*/.*')}`
    }

    /* override */ protected searchValues(): SearchValue[] {
        const queryResults = [...this.queryResults.values()]
        const queryResultRepos = new Set(queryResults.map(({ text }) => text))

        // Include repositories from the user history even if they are not
        // present in the local cache. This happens when the user has visited a
        // repository that they haven't searched for in the fuzzy finder.
        const fromHistory = this.userHistory
            .visitedRepos()
            .filter(repoName => !queryResultRepos.has(repoName))
            .map<PersistableQueryResult>(repoName => ({ text: repoName, url: `/${repoName}` }))
        return [...queryResults, ...fromHistory].map<SearchValue>(({ text, url, stars }) => {
            const formattedRepositoryStarCount = formatRepositoryStarCount(stars)
            const icon = <CodeHostIcon repoName={text} />

            return {
                text,
                url,
                icon: icon ? <span className="mr-1">{icon}</span> : undefined,
                textSuffix:
                    stars && stars > 0 && formattedRepositoryStarCount ? (
                        <span className="mx-1 d-inline-flex align-items-baseline">
                            <SearchResultStar aria-label={`${stars} stars`} className="my-auto" />
                            <span aria-hidden={true}>{formattedRepositoryStarCount}</span>
                        </span>
                    ) : undefined,
                historyRanking: () => this.userHistory.lastAccessedRepo(text),
                ranking: stars,
            }
        })
    }
    /* override */ protected async handleRawQueryPromise(query: string): Promise<PersistableQueryResult[]> {
        const client = this.client || (await getWebGraphQLClient())
        const response = await client.query<FuzzyFinderRepoResult, FuzzyFinderRepoVariables>({
            query: getDocumentNode(FUZZY_REPOS_QUERY),
            variables: { query },
        })
        const repositories = response.data?.search?.results?.repositories || []
        const queryResults: PersistableQueryResult[] = repositories?.map(({ name, stars }) => ({
            text: name,
            url: `/${name}`,
            stars,
        }))
        return queryResults
    }

    private async staleResults(values: PersistableQueryResult[]): Promise<PersistableQueryResult[]> {
        const actualRepos = await this.handleRawQueryPromise(`type:repo (${values.map(({ text }) => text).join('|')})`)
        const isActualRepoName = new Set([...actualRepos.map(({ text }) => text)])
        return values.filter(({ text, stars }) => !isActualRepoName.has(text) || stars === undefined)
    }
}
