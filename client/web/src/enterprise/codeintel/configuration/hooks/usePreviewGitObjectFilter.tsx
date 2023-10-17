import { gql } from '@sourcegraph/http-client'
import type { Maybe } from '@sourcegraph/shared/src/graphql-operations'

import type { PreviewGitObjectFilterResult } from '../../../../graphql-operations'

export interface GitObjectPreviewResult {
    preview: {
        name: string
        rev: string
        committedAt: string
    }[]
    totalCount: number
    totalCountYoungerThanThreshold: number | null
}

export const PREVIEW_GIT_OBJECT_FILTER = gql`
    query PreviewGitObjectFilter(
        $id: ID!
        $type: GitObjectType!
        $pattern: String!
        $countObjectsYoungerThanHours: Int
        $first: Int
    ) {
        node(id: $id) {
            ...RepositoryPreviewGitObjectFilter
        }
    }

    fragment RepositoryPreviewGitObjectFilter on Repository {
        __typename
        previewGitObjectFilter(
            type: $type
            pattern: $pattern
            countObjectsYoungerThanHours: $countObjectsYoungerThanHours
            first: $first
        ) {
            nodes {
                name
                rev
                committedAt
            }
            totalCount
            totalCountYoungerThanThreshold
        }
    }
`

export function convertGitObjectFilterResult(
    preview?: PreviewGitObjectFilterResult
): GitObjectPreviewResult | undefined {
    if (preview) {
        return hasNodeRepositoryType(preview)
            ? {
                  ...preview.node.previewGitObjectFilter,
                  preview: preview.node.previewGitObjectFilter.nodes.map(node => ({
                      name: node.name,
                      rev: node.rev,
                      committedAt: node.committedAt,
                  })),
              }
            : undefined
    }

    return undefined
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
