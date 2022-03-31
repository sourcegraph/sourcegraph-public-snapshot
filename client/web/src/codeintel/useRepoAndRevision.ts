import { QueryResult } from '@apollo/client'

import { useQuery } from '@sourcegraph/http-client'

import { ResolveRepoAndRevisionResult, ResolveRepoAndRevisionVariables } from '../graphql-operations'

import { RESOLVE_REPO_REVISION_QUERY } from './ReferencesPanelQueries'

export const useRepoAndRevision = (repoName: string, revision?: string): QueryResult<ResolveRepoAndRevisionResult> => {
    const result = useQuery<ResolveRepoAndRevisionResult, ResolveRepoAndRevisionVariables>(
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
    return result
}
