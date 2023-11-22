import type { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { memoizeObservable } from '@sourcegraph/common'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { RepoNotFoundError } from '@sourcegraph/shared/src/backend/errors'
import type { PlatformContext } from '@sourcegraph/shared/src/platform/context'

import type {
    FileDiffConnectionFields,
    RepositoryComparisonDiffResult,
    RepositoryComparisonDiffVariables,
} from '../../graphql-operations'

export const queryRepositoryComparisonFileDiffs = memoizeObservable(
    ({
        requestGraphQL,
        ...args
    }: {
        repo: string
        base: string | null
        head: string | null
        first?: number
    } & Pick<PlatformContext, 'requestGraphQL'>): Observable<FileDiffConnectionFields> =>
        requestGraphQL<RepositoryComparisonDiffResult, RepositoryComparisonDiffVariables>({
            request: gql`
                query RepositoryComparisonDiff($repo: String!, $base: String, $head: String, $first: Int) {
                    repository(name: $repo) {
                        comparison(base: $base, head: $head) {
                            fileDiffs(first: $first) {
                                ...FileDiffConnectionFields
                            }
                        }
                    }
                }

                fragment FileDiffConnectionFields on FileDiffConnection {
                    nodes {
                        ...FileDiffFields
                    }
                    totalCount
                }

                fragment FileDiffFields on FileDiff {
                    oldPath
                    newPath
                    internalID
                }
            `,
            variables: { repo: args.repo, base: args.base, head: args.head, first: args.first ?? null },
            mightContainPrivateInfo: true,
        }).pipe(
            map(dataOrThrowErrors),
            map(({ repository }) => {
                if (!repository) {
                    throw new RepoNotFoundError(args.repo)
                }
                if (!repository.comparison?.fileDiffs) {
                    throw new Error('empty fileDiffs')
                }
                return repository.comparison.fileDiffs
            })
        ),
    ({ repo, base, head, first }) => `${repo}:${String(base)}:${String(head)}:${String(first)}`
)
