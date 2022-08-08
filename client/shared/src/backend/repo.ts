import { from, Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { createAggregateError, memoizeObservable } from '@sourcegraph/common'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import {
    TreeEntriesResult,
    TreeFields,
    RepositoryRedirectResult,
    RepositoryRedirectVariables,
    RepositoryFields,
    ResolveRevResult,
    ResolveRevVariables,
} from '../graphql-operations'
import { PlatformContext } from '../platform/context'
import * as GQL from '../schema'
import { AbsoluteRepoFile, makeRepoURI, RepoSpec, RevisionSpec, ResolvedRevisionSpec } from '../util/url'

import { CloneInProgressError, RepoNotFoundError, RepoSeeOtherError, RevisionNotFoundError } from './errors'

/**
 * @returns Observable that emits the `rawRepoName`. Errors with a `CloneInProgressError` if the repo is still being cloned.
 */
export const resolveRawRepoName = memoizeObservable(
    ({
        requestGraphQL,
        repoName,
    }: Pick<RepoSpec, 'repoName'> & Pick<PlatformContext, 'requestGraphQL'>): Observable<string> =>
        from(
            requestGraphQL<GQL.IQuery>({
                request: gql`
                    query ResolveRawRepoName($repoName: String!) {
                        repository(name: $repoName) {
                            uri
                            mirrorInfo {
                                cloned
                            }
                        }
                    }
                `,
                variables: { repoName },
                mightContainPrivateInfo: true,
            })
        ).pipe(
            map(dataOrThrowErrors),
            map(({ repository }) => {
                if (!repository) {
                    throw new RepoNotFoundError(repoName)
                }
                if (!repository.mirrorInfo.cloned) {
                    throw new CloneInProgressError(repoName)
                }
                return repository.uri
            })
        ),
    ({ repoName }) => repoName
)

export const fetchTreeEntries = memoizeObservable(
    ({
        requestGraphQL,
        ...args
    }: AbsoluteRepoFile & { first?: number } & Pick<PlatformContext, 'requestGraphQL'>): Observable<TreeFields> =>
        requestGraphQL<TreeEntriesResult>({
            request: gql`
                query TreeEntries(
                    $repoName: String!
                    $revision: String!
                    $commitID: String!
                    $filePath: String!
                    $first: Int
                ) {
                    repository(name: $repoName) {
                        commit(rev: $commitID, inputRevspec: $revision) {
                            tree(path: $filePath) {
                                ...TreeFields
                            }
                        }
                    }
                }
                fragment TreeFields on GitTree {
                    isRoot
                    url
                    entries(first: $first, recursiveSingleChild: true) {
                        ...TreeEntryFields
                    }
                }
                fragment TreeEntryFields on TreeEntry {
                    name
                    path
                    isDirectory
                    url
                    submodule {
                        url
                        commit
                    }
                    isSingleChild
                }
            `,
            variables: args,
            mightContainPrivateInfo: true,
        }).pipe(
            map(({ data, errors }) => {
                if (errors || !data?.repository?.commit?.tree) {
                    throw createAggregateError(errors)
                }
                return data.repository.commit.tree
            })
        ),
    ({ first, requestGraphQL, ...args }) => `${makeRepoURI(args)}:first-${String(first)}`
)

export const repositoryFragment = gql`
    fragment RepositoryFields on Repository {
        id
        name
        url
        externalURLs {
            url
            serviceKind
        }
        description
        viewerCanAdminister
        defaultBranch {
            displayName
            abbrevName
        }
    }
`

/**
 * Fetch the repository.
 */
export const fetchRepository = memoizeObservable(
    ({
        requestGraphQL,
        ...args
    }: { repoName: string } & Pick<PlatformContext, 'requestGraphQL'>): Observable<RepositoryFields> =>
        requestGraphQL<RepositoryRedirectResult, RepositoryRedirectVariables>({
            request: gql`
                query RepositoryRedirect($repoName: String!) {
                    repositoryRedirect(name: $repoName) {
                        __typename
                        ... on Repository {
                            ...RepositoryFields
                        }
                        ... on Redirect {
                            url
                        }
                    }
                }
                ${repositoryFragment}
            `,
            variables: args,
            mightContainPrivateInfo: true,
        }).pipe(
            map(dataOrThrowErrors),
            map(data => {
                if (!data.repositoryRedirect) {
                    throw new RepoNotFoundError(args.repoName)
                }
                if (data.repositoryRedirect.__typename === 'Redirect') {
                    throw new RepoSeeOtherError(data.repositoryRedirect.url)
                }
                return data.repositoryRedirect
            })
        ),
    makeRepoURI
)

export interface ResolvedRevision extends ResolvedRevisionSpec {
    defaultBranch: string

    /** The URL to the repository root tree at the revision. */
    rootTreeURL: string
}

/**
 * When `revision` is undefined, the default branch is resolved.
 *
 * @returns Observable that emits the commit ID. Errors with a `CloneInProgressError` if the repo is still being cloned.
 */
export const resolveRevision = memoizeObservable(
    ({
        repoName,
        revision,
        requestGraphQL,
    }: RepoSpec & Partial<RevisionSpec> & Pick<PlatformContext, 'requestGraphQL'>): Observable<ResolvedRevision> =>
        requestGraphQL<ResolveRevResult, ResolveRevVariables>({
            request: gql`
                query ResolveRev($repoName: String!, $revision: String!) {
                    repositoryRedirect(name: $repoName) {
                        __typename
                        ... on Repository {
                            mirrorInfo {
                                cloneInProgress
                                cloneProgress
                                cloned
                            }
                            commit(rev: $revision) {
                                oid
                                tree(path: "") {
                                    url
                                }
                            }
                            defaultBranch {
                                abbrevName
                            }
                        }
                        ... on Redirect {
                            url
                        }
                    }
                }
            `,
            variables: { repoName, revision: revision || '' },
            mightContainPrivateInfo: true,
        }).pipe(
            map(({ data, errors }) => {
                if (!data) {
                    throw createAggregateError(errors)
                }
                if (!data.repositoryRedirect) {
                    throw new RepoNotFoundError(repoName)
                }
                if (data.repositoryRedirect.__typename === 'Redirect') {
                    throw new RepoSeeOtherError(data.repositoryRedirect.url)
                }
                if (data.repositoryRedirect.mirrorInfo.cloneInProgress) {
                    throw new CloneInProgressError(
                        repoName,
                        data.repositoryRedirect.mirrorInfo.cloneProgress || undefined
                    )
                }
                if (!data.repositoryRedirect.mirrorInfo.cloned) {
                    throw new CloneInProgressError(repoName, 'queued for cloning')
                }
                if (!data.repositoryRedirect.commit) {
                    throw new RevisionNotFoundError(revision)
                }

                const defaultBranch = data.repositoryRedirect.defaultBranch?.abbrevName || 'HEAD'

                if (!data.repositoryRedirect.commit.tree) {
                    throw new RevisionNotFoundError(defaultBranch)
                }
                return {
                    commitID: data.repositoryRedirect.commit.oid,
                    defaultBranch,
                    rootTreeURL: data.repositoryRedirect.commit.tree.url,
                }
            })
        ),
    makeRepoURI
)
