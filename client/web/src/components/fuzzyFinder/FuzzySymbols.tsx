import { ApolloClient } from '@apollo/client'
import { FuzzyFinderSymbolsResult, FuzzyFinderSymbolsVariables } from 'src/graphql-operations'
import gql from 'tagged-template-noop'

import { getDocumentNode } from '@sourcegraph/http-client'
import { SymbolTag } from '@sourcegraph/shared/src/symbols/SymbolTag'

import { getWebGraphQLClient } from '../../backend/graphql'
import { SearchValue } from '../../fuzzyFinder/FuzzySearch'

import { emptyFuzzyCache, PersistableQueryResult } from './FuzzyLocalCache'
import { FuzzyQuery } from './FuzzyQuery'
import { FuzzyRepoRevision, fuzzyRepoRevisionSearchFilter } from './FuzzyRepoRevision'

export const FUZZY_SYMBOLS_QUERY = gql`
    fragment FileMatchFields on FileMatch {
        symbols {
            name
            containerName
            kind
            language
            url
        }
        repository {
            name
        }
        file {
            path
        }
    }

    query FuzzyFinderSymbols($query: String!) {
        search(patternType: regexp, query: $query) {
            results {
                results {
                    ... on FileMatch {
                        __typename
                        ...FileMatchFields
                    }
                }
            }
        }
    }
`

export class FuzzySymbols extends FuzzyQuery {
    constructor(
        private readonly client: ApolloClient<object> | undefined,
        onNamesChanged: () => void,
        private readonly repoRevision: React.MutableRefObject<FuzzyRepoRevision>,
        private readonly isGlobalSymbolsRef: React.MutableRefObject<boolean>
    ) {
        // Symbol results should not be cached because stale symbol data is complicated to evict/invalidate.
        super(onNamesChanged, emptyFuzzyCache)
    }

    /* override */ protected searchValues(): SearchValue[] {
        const repositoryName = this.repoRevision.current.repositoryName
        const repositoryFilter = repositoryName && !this.isGlobalSymbolsRef.current ? '/' + repositoryName : ''
        let values = [...this.queryResults.values()]
        if (repositoryFilter) {
            values = values.filter(({ url }) => url?.startsWith(repositoryFilter))
        }

        const repositoryText = `${repositoryName}/`
        return values.map(({ text, url, symbolKind }) => ({
            text: repositoryFilter ? text.replace(repositoryText, '') : text,
            url,
            icon: symbolKind ? <SymbolTag kind={symbolKind} /> : undefined,
        }))
    }

    /* override */ protected rawQuery(query: string): string {
        const repoFilter = this.isGlobalSymbolsRef.current
            ? ''
            : fuzzyRepoRevisionSearchFilter(this.repoRevision.current)
        return `${repoFilter}type:symbol count:10 ${query}`
    }

    /* override */ protected async handleRawQueryPromise(query: string): Promise<PersistableQueryResult[]> {
        const client = this.client || (await getWebGraphQLClient())
        const response = await client.query<FuzzyFinderSymbolsResult, FuzzyFinderSymbolsVariables>({
            query: getDocumentNode(FUZZY_SYMBOLS_QUERY),
            variables: { query },
        })
        const results = response.data?.search?.results?.results || []
        const queryResults: PersistableQueryResult[] = []
        for (const result of results) {
            if (result.__typename === 'FileMatch') {
                for (const symbol of result.symbols) {
                    const repository = result.repository.name ? `${result.repository.name}/` : ''
                    const containerName = symbol.containerName ? ` (${symbol.containerName})` : ''
                    queryResults.push({
                        text: `${symbol.name}${containerName} - ${repository}${result.file.path} - ${symbol.language}`,
                        url: symbol.url,
                        symbolKind: symbol.kind,
                    })
                }
            }
        }
        return queryResults
    }
}
