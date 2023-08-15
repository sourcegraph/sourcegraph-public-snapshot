import type { ErrorLike } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'

import type { ResolveRepoAndRevisionResult, ResolveRepoAndRevisionVariables } from '../graphql-operations'

import { RESOLVE_REPO_REVISION_BLOB_QUERY } from './ReferencesPanelQueries'

interface RepoAndBlob {
    isFork: boolean
    isArchived: boolean
    revision: string
    commitID: string
    fileContent: string
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

    if (loading || error) {
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
        },
    }
}
