import { Fzf, type FzfOptions, type FzfResultItem } from 'fzf'

import { gql } from '@sourcegraph/http-client'

import type { SuggestionsRepoResult, SuggestionsRepoVariables } from '../../graphql-operations'

import { CachedAsyncCompletionSource } from './source'

interface Repo {
    name: string
    stars: number
}

const REPOS_QUERY = gql`
    query SuggestionsRepo($query: String!) {
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

type RequestGraphQL<R, V> = (query: string, variables: V) => Promise<R>

function starTiebraker(a: { item: { stars: number } }, b: { item: { stars: number } }): number {
    return b.item.stars - a.item.stars
}

const repoFzfOptions: FzfOptions<Repo> = {
    selector: item => item.name,
    tiebreakers: [starTiebraker],
    forward: false,
}

const cleanRegex = (value: string): string => value.replaceAll(/^\^|\\\.|\$$/g, '')

export function createRepositoryCompletionSource<T, ExtraArgs extends any[]>(
    request: RequestGraphQL<SuggestionsRepoResult, SuggestionsRepoVariables>,
    dataCacheKey?: (...args: ExtraArgs) => string
): CachedAsyncCompletionSource<Repo, FzfResultItem<Repo>, ExtraArgs> {
    return new CachedAsyncCompletionSource({
        dataCacheKey,
        queryKey: (value, dataCacheKey = '') => `${dataCacheKey} type:repo count:50 repo:${value}`,
        async query(query) {
            const response = await request(REPOS_QUERY, { query })
            return response.search?.results?.repositories.map(repository => [repository.name, repository]) || []
        },
        filter(repos, query) {
            const fzf = new Fzf(repos, repoFzfOptions)
            return fzf.find(cleanRegex(query))
        },
    })
}
