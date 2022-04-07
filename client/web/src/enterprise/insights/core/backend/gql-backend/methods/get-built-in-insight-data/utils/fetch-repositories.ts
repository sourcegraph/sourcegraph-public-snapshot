import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import { requestGraphQL } from '../../../../../../../../backend/graphql'
import { BulkSearchRepositories } from '../../../../../../../../graphql-operations'

const bulkSearchRepositoriesFragment = gql`
    fragment BulkSearchRepositories on Repository {
        name
    }
`

export function fetchRepositories(repositories: string[]): Observable<BulkSearchRepositories[]> {
    return requestGraphQL<Record<string, BulkSearchRepositories>>(
        `
        query BulkRepositoriesSearch(${repositories.map((repo, index) => `$query${index}: String!`).join(', ')}) {
            ${repositories
                .map(
                    (_repo, index) => `
                    repoSearch${index}: repositoryRedirect(name: $query${index}) {
                        ...BulkSearchRepositories
                    }
                `
                )
                .join('\n')}
        }
        ${bulkSearchRepositoriesFragment}
    `,
        Object.fromEntries(repositories.map((query, index) => [`query${index}`, query]))
    ).pipe(
        map(dataOrThrowErrors),
        map(search =>
            /**
             * Gather information from bulk search to array of search results
             *
             * Raw search:
             * { repoSearch0: { repo content 1 }, repoSearch1: { repo content 2 } ... }
             *
             * Transformed array result
             *
             * [{ repo content 1}, { repo content 2 }, ... { repo content N }]
             * */
            Object.keys(search).reduce<BulkSearchRepositories[]>((result, key) => {
                const index = +key.slice('repoSearch'.length)

                result[index] = search[key] ?? {}

                return result
            }, [])
        )
    )
}
