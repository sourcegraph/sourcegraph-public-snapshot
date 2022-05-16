import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { createAggregateError, memoizeObservable } from '@sourcegraph/common'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import {
    CloneInProgressError,
    RepoNotFoundError,
    RepoSeeOtherError,
    RevisionNotFoundError,
} from '@sourcegraph/shared/src/backend/errors'
import {
    makeRepoURI,
    RepoRevision,
    RevisionSpec,
    RepoSpec,
    ResolvedRevisionSpec,
} from '@sourcegraph/shared/src/util/url'

import { queryGraphQL, requestGraphQL } from '../backend/graphql'
import {
    ExternalLinkFields,
    RepositoryRedirectResult,
    RepositoryRedirectVariables,
    RepositoryFields,
} from '../graphql-operations'

export const externalLinkFieldsFragment = gql`
    fragment ExternalLinkFields on ExternalLink {
        url
        serviceKind
    }
`

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
    (args: { repoName: string }): Observable<RepositoryFields> =>
        requestGraphQL<RepositoryRedirectResult, RepositoryRedirectVariables>(
            gql`
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
            args
        ).pipe(
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
    ({ repoName, revision }: RepoSpec & Partial<RevisionSpec>): Observable<ResolvedRevision> =>
        queryGraphQL(
            gql`
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
            { repoName, revision: revision || '' }
        ).pipe(
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

export const fetchFileExternalLinks = memoizeObservable(
    (context: RepoRevision & { filePath: string }): Observable<ExternalLinkFields[]> =>
        queryGraphQL(
            gql`
                query FileExternalLinks($repoName: String!, $revision: String!, $filePath: String!) {
                    repository(name: $repoName) {
                        commit(rev: $revision) {
                            file(path: $filePath) {
                                externalURLs {
                                    ...ExternalLinkFields
                                }
                            }
                        }
                    }
                }

                ${externalLinkFieldsFragment}
            `,
            context
        ).pipe(
            map(({ data, errors }) => {
                if (!data?.repository?.commit?.file?.externalURLs) {
                    throw createAggregateError(errors)
                }
                return data.repository.commit.file.externalURLs
            })
        ),
    makeRepoURI
)
