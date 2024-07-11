import { error } from '@sveltejs/kit'

import { isErrorLike } from '$lib/common'
import { getGraphQLClient, mapOrThrow } from '$lib/graphql'
import { GitRefType } from '$lib/graphql-types'

import type { LayoutLoad } from './$types'
import { RepositoryGitCommits, RepositoryGitRefs } from './layout.gql'

export const load: LayoutLoad = async ({ parent }) => {
    // By validating the resolved revision here we can guarantee to
    // subpages that if they load the requested revision exists. This
    // relieves subpages from testing whether the revision is valid.
    const { repoName, resolvedRevisionOrError } = await parent()

    if (isErrorLike(resolvedRevisionOrError)) {
        error(404, resolvedRevisionOrError)
    }

    return {
        resolvedRevision: resolvedRevisionOrError,
        // Repository pickers queries (branch, tags and commits)
        getRepoBranches: (searchTerm: string) =>
            getGraphQLClient()
                .query(RepositoryGitRefs, {
                    repoName: repoName,
                    query: searchTerm,
                    type: GitRefType.GIT_BRANCH,
                })
                .then(
                    mapOrThrow(({ data, error }) => {
                        if (!data?.repository?.gitRefs) {
                            throw new Error(error?.message)
                        }

                        return data.repository.gitRefs
                    })
                ),
        getRepoTags: (searchTerm: string) =>
            getGraphQLClient()
                .query(RepositoryGitRefs, {
                    repoName,
                    query: searchTerm,
                    type: GitRefType.GIT_TAG,
                })
                .then(
                    mapOrThrow(({ data, error }) => {
                        if (!data?.repository?.gitRefs) {
                            throw new Error(error?.message)
                        }

                        return data.repository.gitRefs
                    })
                ),
        getRepoCommits: (searchTerm: string) =>
            getGraphQLClient()
                .query(RepositoryGitCommits, {
                    repoName,
                    query: searchTerm,
                    revision: resolvedRevisionOrError.commitID,
                })
                .then(
                    mapOrThrow(({ data }) => {
                        let nodes = data?.repository?.ancestorCommits?.ancestors.nodes ?? []

                        // If we got a match for the OID, add it to the list if it doesn't already exist.
                        // We double check that the OID contains the search term because we cannot search
                        // specifically by OID, and an empty string resolves to HEAD.
                        const commitByHash = data?.repository?.commitByHash
                        if (
                            commitByHash &&
                            commitByHash.oid.includes(searchTerm) &&
                            !nodes.some(node => node.oid === commitByHash.oid)
                        ) {
                            nodes = [commitByHash, ...nodes]
                        }
                        return { nodes }
                    })
                ),
    }
}
