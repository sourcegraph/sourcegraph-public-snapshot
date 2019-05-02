import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContext } from '../../../../../shared/src/platform/context'
import { memoizeObservable } from '../../../../../shared/src/util/memoizeObservable'
import { isPrivateRepository } from '../util/context'
import { RepoNotFoundError } from './errors'

export const queryRepositoryComparisonFileDiffs = memoizeObservable(
    ({
        queryGraphQL,
        ...args
    }: {
        repo: string
        base: string | null
        head: string | null
        first?: number
        queryGraphQL: PlatformContext['requestGraphQL']
    }): Observable<GQL.IFileDiffConnection> =>
        queryGraphQL<GQL.IQuery>(
            gql`
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
            { repo: args.repo, base: args.base, head: args.head, first: args.first },
            // This request may contain private info if the repository is private
            isPrivateRepository()
        ).pipe(
            map(dataOrThrowErrors),
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
