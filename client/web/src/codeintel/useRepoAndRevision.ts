import { ErrorLike } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'

import { ResolveRepoAndRevisionResult, ResolveRepoAndRevisionVariables } from '../graphql-operations'

import { RESOLVE_REPO_REVISION_QUERY } from './ReferencesPanelQueries'

interface RepoAndRevision {
    isFork: boolean
    isArchived: boolean
    revision: string
    commitID: string
}
interface UseRepoAndRevisionResult {
    loading: boolean
    error?: ErrorLike

    data?: RepoAndRevision
}

export const useRepoAndRevision = (repoName: string, revision?: string): UseRepoAndRevisionResult => {
    const { data, loading, error } = useQuery<ResolveRepoAndRevisionResult, ResolveRepoAndRevisionVariables>(
        RESOLVE_REPO_REVISION_QUERY,
        {
            variables: {
                repoName,
                revision: revision || '',
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

    return {
        loading: false,
        data: {
            isArchived: repository.isArchived,
            isFork: repository.isFork,
            revision: revision ?? defaultBranch,
            commitID: repository.commit.oid,
        },
    }
}
