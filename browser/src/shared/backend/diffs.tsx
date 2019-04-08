import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import * as GQL from '../../../../shared/src/graphql/schema'
import { memoizeObservable } from '../../../../shared/src/util/memoizeObservable'
import { getContext } from './context'
import { RepoNotFoundError } from './errors'
import { queryGraphQL } from './graphql'

export const queryRepositoryComparisonFileDiffs = memoizeObservable(
    (args: {
        repo: string
        base: string | null
        head: string | null
        first?: number
    }): Observable<GQL.IFileDiffConnection> =>
        queryGraphQL({
            ctx: getContext(),
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
            map(({ repository }) => {
                if (!repository) {
                    throw new RepoNotFoundError(args.repo)
                }
                if (!repository.comparison || !repository.comparison.fileDiffs) {
                    throw new Error('empty fileDiffs')
                }
                return repository.comparison.fileDiffs
            })
        ),
    ({ repo, base, head, first }) => `${repo}:${base}:${head}:${first}`
)
