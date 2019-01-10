import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { createAggregateError } from './errors'
import { queryGraphQL } from './graphql'

export function queryRepositoryComparisonFileDiffs(args: {
    repo: string
    base: string | null
    head: string | null
    first?: number
}): Observable<GQL.IFileDiffConnection> {
    console.log('scalar ID/repo', args.repo)
    return queryGraphQL({
        ctx: { repoKey: '', isRepoSpecific: false },
        request: `
            query RepositoryComparisonDiff($repo: String!, $base: String, $head: String, $first: Int) {
                repository(name: $repo) {
                    ... on Repository {
                        comparison(base: $base, head: $head) {
                            fileDiffs(first: $first) {
                                nodes {
                                    ...FileDiffFields
                                }
                                totalCount
                                pageInfo {
                                    hasNextPage
                                }
                            }
                        }
                    }
                }
            }

            fragment FileDiffFields on FileDiff {
                oldPath
                newPath
                mostRelevantFile {
                    url
                }
                hunks {
                    oldNoNewlineAt
                    section
                    body
                }
                internalID
            }
        `,
        variables: { repo: args.repo, base: args.base, head: args.head, first: args.first },
    }).pipe(
        map(({ data, errors }) => {
            console.log('DATA', data)
            if (!data || !data.repository) {
                throw createAggregateError(errors)
            }
            const repo = data.repository as GQL.IRepository
            if (!repo.comparison || !repo.comparison.fileDiffs || errors) {
                throw createAggregateError(errors)
            }
            return repo.comparison.fileDiffs
        })
    )
}
