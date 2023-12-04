import { query, gql, fromCache } from '$lib/graphql'
import type { ResolveRepoRevisonResult, ResolveRepoRevisonVariables } from '$lib/graphql-operations'
import {
    CloneInProgressError,
    RepoNotFoundError,
    RepoSeeOtherError,
    RevisionNotFoundError,
    type RepoSpec,
    type RevisionSpec,
    type ResolvedRevisionSpec,
} from '$lib/shared'

const resolveRevisionQuery = gql`
    query ResolveRepoRevison($repoName: String!, $revision: String!) {
        repositoryRedirect(name: $repoName) {
            __typename
            ... on Repository {
                id
                name
                url
                sourceType
                externalURLs {
                    url
                    serviceKind
                }
                externalRepository {
                    serviceType
                    serviceID
                }
                description
                viewerCanAdminister
                defaultBranch {
                    displayName
                    abbrevName
                }
                isFork
                metadata {
                    key
                    value
                }
                mirrorInfo {
                    cloneInProgress
                    cloneProgress
                    cloned
                }
                commit(rev: $revision) {
                    id
                    oid
                    tree(path: ".") {
                        canonicalURL
                        url
                    }
                }
                changelist(cid: $revision) {
                    cid
                    canonicalURL
                    commit {
                        id
                        __typename
                        oid
                        tree(path: ".") {
                            canonicalURL
                            url
                        }
                    }
                }
                defaultBranch {
                    id
                    abbrevName
                }
            }
            ... on Redirect {
                url
            }
        }
    }
`

export interface ResolvedRevision extends ResolvedRevisionSpec {
    defaultBranch: string
    repo: Extract<ResolveRepoRevisonResult['repositoryRedirect'], { __typename: 'Repository' }>
}

/**
 * When `revision` is undefined, the default branch is resolved.
 *
 * @returns Promise that emits the commit ID. Errors with a `CloneInProgressError` if the repo is still being cloned.
 */
export async function resolveRepoRevision({
    repoName,
    revision = '',
}: RepoSpec & Partial<RevisionSpec>): Promise<ResolvedRevision> {
    let data = await fromCache<ResolveRepoRevisonResult, ResolveRepoRevisonVariables>(resolveRevisionQuery, {
        repoName,
        revision,
    })
    if (
        !data ||
        (data.repositoryRedirect?.__typename === 'Repository' && data.repositoryRedirect.commit?.oid !== revision)
    ) {
        // We always refetch data when 'revision' is a "symbolic" revision (e.g. a tag or branch name)
        data = await query<ResolveRepoRevisonResult, ResolveRepoRevisonVariables>(
            resolveRevisionQuery,
            {
                repoName,
                revision,
            },
            { fetchPolicy: 'network-only' }
        )
    }

    if (!data.repositoryRedirect) {
        throw new RepoNotFoundError(repoName)
    }
    if (data.repositoryRedirect.__typename === 'Redirect') {
        throw new RepoSeeOtherError(data.repositoryRedirect.url)
    }
    if (data.repositoryRedirect.mirrorInfo.cloneInProgress) {
        throw new CloneInProgressError(repoName, data.repositoryRedirect.mirrorInfo.cloneProgress || undefined)
    }
    if (!data.repositoryRedirect.mirrorInfo.cloned) {
        throw new CloneInProgressError(repoName, 'queued for cloning')
    }

    // The "revision" we queried for could be a commit or a changelist.
    const commit = data.repositoryRedirect.commit || data.repositoryRedirect.changelist?.commit
    if (!commit) {
        throw new RevisionNotFoundError(revision)
    }

    const defaultBranch = data.repositoryRedirect.defaultBranch?.abbrevName || 'HEAD'

    if (!commit.tree) {
        throw new RevisionNotFoundError(defaultBranch)
    }

    return {
        repo: data.repositoryRedirect,
        commitID: commit.oid,
        defaultBranch,
    }
}
