import { ApolloClient } from '@apollo/client'
import gql from 'tagged-template-noop'

import { getDocumentNode } from '@sourcegraph/http-client'
import { CodeHostIcon } from '@sourcegraph/search-ui'

import { getWebGraphQLClient } from '../../backend/graphql'
import { SearchValue } from '../../fuzzyFinder/FuzzySearch'
import { FuzzyFinderRepoResult, FuzzyFinderRepoVariables } from '../../graphql-operations'

import { FuzzyWebCache, PersistableQueryResult } from './FuzzyLocalCache'
import { FuzzyQuery } from './FuzzyQuery'

export const FUZZY_REPOS_QUERY = gql`
    query FuzzyFinderRepo($query: String!) {
        search(patternType: regexp, query: $query) {
            results {
                repositories {
                    name
                }
            }
        }
    }
`

export class FuzzyRepos extends FuzzyQuery {
    constructor(private readonly client: ApolloClient<object> | undefined, onNamesChanged: () => void) {
        super(onNamesChanged, new FuzzyWebCache('fuzzy-finder.repository-names', values => this.staleResults(values)))
    }

    /* override */ protected rawQuery(query: string): string {
        return `type:repo count:10 ${query.replace('/', '.*/.*')}`
    }

    /* override */ protected searchValues(): SearchValue[] {
        return [...this.queryResults.values()].map(({ text, url }) => ({
            text,
            url,
            icon: <CodeHostIcon repoName={text} /> || undefined,
        }))
    }
    /* override */ protected async handleRawQueryPromise(query: string): Promise<PersistableQueryResult[]> {
        const client = this.client || (await getWebGraphQLClient())
        const response = await client.query<FuzzyFinderRepoResult, FuzzyFinderRepoVariables>({
            query: getDocumentNode(FUZZY_REPOS_QUERY),
            variables: { query },
        })
        const repositories = response.data?.search?.results?.repositories || []
        // const repositories = await Promise.resolve([{ name: 'a' }])
        return repositories?.map(({ name }) => ({
            text: name,
            url: `/${name}`,
        }))
    }

    private async staleResults(values: PersistableQueryResult[]): Promise<PersistableQueryResult[]> {
        const actualRepos = await this.handleRawQueryPromise(`type:repo (${values.map(({ text }) => text).join('|')})`)
        const isActualRepoName = new Set([...actualRepos.map(({ text }) => text)])
        return values.filter(({ text }) => !isActualRepoName.has(text))
    }
}
