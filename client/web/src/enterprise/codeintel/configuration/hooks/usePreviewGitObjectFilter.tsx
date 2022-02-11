import { ApolloError } from '@apollo/client'

import { gql, useQuery } from '@sourcegraph/http-client'
import { GitObjectType, Maybe } from '@sourcegraph/shared/src/graphql-operations'

import { PreviewGitObjectFilterResult, PreviewGitObjectFilterVariables } from '../../../../graphql-operations'

interface SearchGitObjectResult {
    previewResult: GitObjectPreviewResult
    isLoadingPreview: boolean
    previewError: ApolloError | undefined
}

export interface GitObjectPreviewResult {
    preview: {
        repoName: string
        name: string
        rev: string
    }[]
}

export const PREVIEW_GIT_OBJECT_FILTER = gql`
    query PreviewGitObjectFilter($id: ID!, $type: GitObjectType!, $pattern: String!) {
        node(id: $id) {
            ...RepositoryPreviewGitObjectFilter
        }
    }

    fragment RepositoryPreviewGitObjectFilter on Repository {
        __typename
        name
        previewGitObjectFilter(type: $type, pattern: $pattern) {
            name
            rev
        }
    }
`

export const usePreviewGitObjectFilter = (id: string, type: GitObjectType, pattern: string): SearchGitObjectResult => {
    const { data, loading, error } = useQuery<PreviewGitObjectFilterResult, PreviewGitObjectFilterVariables>(
        PREVIEW_GIT_OBJECT_FILTER,
        {
            variables: { id, type, pattern },
        }
    )

    return {
        previewResult: {
            preview: hasNodeRepositoryType(data)
                ? data.node.previewGitObjectFilter.map(({ name, rev }) => ({ repoName: data.node.name, name, rev }))
                : [],
        },
        isLoadingPreview: loading,
        previewError: error,
    }
}

function hasNodeRepositoryType<
    T extends {
        node: Maybe<{
            __typename?: string | 'Repository'
        }>
    }
>(
    data: T | undefined
): data is T & {
    node: {
        __typename: 'Repository'
    }
} {
    return data?.node?.__typename === 'Repository'
}
