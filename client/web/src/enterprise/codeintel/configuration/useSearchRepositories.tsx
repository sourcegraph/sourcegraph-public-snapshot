import { ApolloError } from '@apollo/client'

import { gql, useQuery } from '@sourcegraph/shared/src/graphql/graphql'
import { PreviewRepositoryFilterResult, PreviewRepositoryFilterVariables } from '../../../graphql-operations'

interface SearchRepositoriesResult {
    previewResult: RepositoryPreviewResult
    isLoadingPreview: boolean
    previewError: ApolloError | undefined
}

export interface RepositoryPreviewResult {
    preview: {
        name: string
    }[]
}

export const PREVIEW_REPOSITORY_FILTER = gql`
    query PreviewRepositoryFilter($pattern: String!) {
        previewRepositoryFilter(pattern: $pattern) {
            name
        }
    }
`

export const usePreviewRepositoryFilter = (pattern: string): SearchRepositoriesResult => {
    const { data, loading, error } = useQuery<PreviewRepositoryFilterResult, PreviewRepositoryFilterVariables>(
        PREVIEW_REPOSITORY_FILTER,
        {
            variables: {
                pattern,
            },
        }
    )

    return {
        previewResult: {
            preview: data?.previewRepositoryFilter.map(({ name }) => ({ name })) || [],
        },
        isLoadingPreview: loading,
        previewError: error,
    }
}
