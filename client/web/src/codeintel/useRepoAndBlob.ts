import type { ErrorLike } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'

import {
    HighlightResponseFormat,
    type ReferencesPanelHighlightedBlobResult,
    type ReferencesPanelHighlightedBlobVariables,
    type ResolveRepoAndRevisionResult,
    type ResolveRepoAndRevisionVariables,
} from '../graphql-operations'

import { FETCH_HIGHLIGHTED_BLOB, RESOLVE_REPO_REVISION_BLOB_QUERY } from './ReferencesPanelQueries'
import { Occurrence } from '@sourcegraph/shared/src/codeintel/scip'

interface RepoAndBlob {
    isFork: boolean
    isArchived: boolean
    revision: string
    commitID: string
    fileContent: string
    occurrences: Occurrence[]
}

interface UseRepoAndBlobResult {
    loading: boolean
    error?: ErrorLike

    data?: RepoAndBlob
}

export const useRepoAndBlob = (repoName: string, filePath: string, revision?: string): UseRepoAndBlobResult => {
    const { data, loading, error } = useQuery<ResolveRepoAndRevisionResult, ResolveRepoAndRevisionVariables>(
        RESOLVE_REPO_REVISION_BLOB_QUERY,
        {
            variables: {
                repoName,
                revision: revision || '',
                filePath,
            },
            notifyOnNetworkStatusChange: false,
            fetchPolicy: 'no-cache',
        }
    )

    const { data: highlightingData, error: highlightingError, loading: highlightingLoading }  = useQuery<
        ReferencesPanelHighlightedBlobResult,
        ReferencesPanelHighlightedBlobVariables
    >(FETCH_HIGHLIGHTED_BLOB, {
        variables: {
            repository: repoName,
            commit: revision ?? 'BOOM',
            path: filePath,
            format: HighlightResponseFormat.JSON_SCIP,
        },
        // Cache this data but always re-request it in the background when we revisit
        // this page to pick up newer changes.
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'network-only',
    })
    const lsif = highlightingData?.repository?.commit?.blob?.highlight?.lsif
    console.log({lsif})

    if (loading || error || highlightingError || highlightingLoading) {
        return { loading, error }
    }

    if (!data) {
        return { loading }
    }

    if (data?.repositoryRedirect?.__typename !== 'Repository') {
        return { loading }
    }

    const repository = data.repositoryRedirect
    const defaultBranch = repository.defaultBranch?.abbrevName || 'HEAD'

    if (!repository.commit) {
        return { loading, error: { message: `revision not found: ${defaultBranch}` } }
    }

    if (!repository.commit.file) {
        return { loading, error: { message: `file not found: ${filePath}` } }
    }

    return {
        loading: false,
        data: {
            isArchived: repository.isArchived,
            isFork: repository.isFork,
            revision: revision ?? defaultBranch,
            commitID: repository.commit.oid,
            fileContent: repository.commit.file.content,
            occurrences: [].
        },
    }
}
