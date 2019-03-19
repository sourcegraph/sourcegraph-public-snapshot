import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { memoizeObservable } from '../../../../../shared/src/util/memoizeObservable'
import { createAggregateError } from './errors'
import { queryGraphQL } from './graphql'

export const queryRepositoryComparisonFileDiffs = memoizeObservable(
    (args: {
        repo: string
        base: string | null
        head: string | null
        first?: number
    }): Observable<GQL.IFileDiffConnection> =>
        queryGraphQL({
            ctx: { repoKey: '', isRepoSpecific: false },
            request: `
            query RepositoryComparisonDiff($repo: String!, $base: String, $head: String, $first: Int) {
                repository(name: $repo) {
                    comparison(base: $base, head: $head) {
                        fileDiffs(first: $first) {
                            nodes {
                                ...FileDiffFields
                            }
                            totalCount
                        }
                    }
                }
            }

            fragment FileDiffFields on FileDiff {
                oldPath
                newPath
                internalID
            }
        `,
            variables: { repo: args.repo, base: args.base, head: args.head, first: args.first },
        }).pipe(
            map(({ data, errors }) => {
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
)
