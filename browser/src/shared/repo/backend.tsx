import { from, Observable } from 'rxjs'
import { catchError, delay, filter, map, retryWhen } from 'rxjs/operators'
import {
    AggregateError,
    CloneInProgressError,
    ECLONEINPROGESS,
    RepoNotFoundError,
    RevNotFoundError,
} from '../../../../shared/src/backend/errors'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { PlatformContext } from '../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../shared/src/util/errors'
import { memoizeObservable } from '../../../../shared/src/util/memoizeObservable'
import { FileSpec, makeRepoURI, RawRepoSpec, RepoSpec, ResolvedRevSpec, RevSpec } from '../../../../shared/src/util/url'

/**
 * @returns Observable that emits if the repo exists on the instance.
 *         Emits the repo name on the Sourcegraph instance as affected by `repositoryPathPattern`.
 *         Errors with a `RepoNotFoundError` if the repo is not found
 */
export const resolveRepo = memoizeObservable(
    ({ rawRepoName, requestGraphQL }: RawRepoSpec & Pick<PlatformContext, 'requestGraphQL'>): Observable<string> =>
        requestGraphQL<GQL.IQuery>({
            request: gql`
                query ResolveRepo($rawRepoName: String!) {
                    repository(name: $rawRepoName) {
                        name
                    }
                }
            `,
            variables: { rawRepoName },
            // This request may leak private repository names
            mightContainPrivateInfo: true,
        }).pipe(
            map(dataOrThrowErrors),
            map(
                ({ repository }) => {
                    if (!repository || !repository.name) {
                        throw new RepoNotFoundError(rawRepoName)
                    }
                    return repository.name
                },
                catchError((err, caught) => caught)
            )
        ),
    ({ rawRepoName }) => rawRepoName
)

/**
 * @returns Observable that emits the commit ID. Errors with a `CloneInProgressError` if the repo is still being cloned.
 */
export const resolveRev = memoizeObservable(
    ({
        requestGraphQL,
        ...ctx
    }: RepoSpec & Partial<RevSpec> & Pick<PlatformContext, 'requestGraphQL'>): Observable<string> =>
        from(
            requestGraphQL<GQL.IQuery>({
                request: gql`
                    query ResolveRev($repoName: String!, $rev: String!) {
                        repository(name: $repoName) {
                            mirrorInfo {
                                cloned
                            }
                            commit(rev: $rev) {
                                oid
                            }
                        }
                    }
                `,
                variables: { ...ctx, rev: ctx.rev || '' },
                mightContainPrivateInfo: true,
            })
        ).pipe(
            map(dataOrThrowErrors),
            map(({ repository }) => {
                if (!repository) {
                    throw new RepoNotFoundError(ctx.repoName)
                }
                if (!repository.mirrorInfo.cloned) {
                    throw new CloneInProgressError(ctx.repoName)
                }
                if (!repository.commit) {
                    throw new RevNotFoundError(ctx.rev)
                }
                return repository.commit.oid
            })
        ),
    makeRepoURI
)

export function retryWhenCloneInProgressError<T>(): (v: Observable<T>) => Observable<T> {
    return (maybeErrors: Observable<T>) =>
        maybeErrors.pipe(
            retryWhen(errors =>
                errors.pipe(
                    filter(err => {
                        if (isErrorLike(err) && err.code === ECLONEINPROGESS) {
                            return true
                        }

                        // Don't swallow other errors.
                        throw err
                    }),
                    delay(1000)
                )
            )
        )
}

/**
 * Fetches the lines of a given file at a given commit from the Sourcegraph API.
 * Will return an empty array if the repo, commit or file does not exist or an error happened (TODO change this!).
 *
 * Only emits once.
 */
export const fetchBlobContentLines = memoizeObservable(
    ({
        requestGraphQL,
        ...ctx
    }: RepoSpec & ResolvedRevSpec & FileSpec & Pick<PlatformContext, 'requestGraphQL'>): Observable<string[]> =>
        from(
            requestGraphQL<GQL.IQuery>({
                request: gql`
                    query BlobContent($repoName: String!, $commitID: String!, $filePath: String!) {
                        repository(name: $repoName) {
                            commit(rev: $commitID) {
                                file(path: $filePath) {
                                    content
                                }
                            }
                        }
                    }
                `,
                variables: ctx,
                mightContainPrivateInfo: true,
            })
        ).pipe(
            map(({ data, errors }) => {
                if (!data) {
                    throw new Error('Invalid response')
                }
                if (errors) {
                    if (errors.length === 1) {
                        const err = errors[0]
                        const isFileContent = err.path.join('.') === 'repository.commit.file.content'
                        const isDNE = err.message.includes('does not exist')

                        // The error is the file DNE. Just ignore it and pass an empty array
                        // to represent this.
                        if (isFileContent && isDNE) {
                            return []
                        }
                    }
                    throw new AggregateError(errors)
                }
                const { repository } = data
                if (!repository || !repository.commit || !repository.commit.file || !repository.commit.file.content) {
                    return []
                }
                return repository.commit.file.content.split('\n')
            })
        ),
    makeRepoURI
)
