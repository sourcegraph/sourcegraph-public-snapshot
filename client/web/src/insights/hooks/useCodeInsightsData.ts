import { QueryResult } from '@apollo/client'

import { gql, useQuery } from '@sourcegraph/http-client'

import { GetRelatedInsightsInlineResult } from '../../graphql-operations'

const RELATED_INSIGHTS_INLINE_QUERY = gql`
    query GetRelatedInsightsInline($input: RelatedInsightsInput!) {
        relatedInsightsInline(input: $input) {
            viewId
            title
            lineNumbers
            text
        }
    }
`

interface UseCodeInsightsDataInput {
    file: string
    revision: string
    repo: string
}

export const useCodeInsightsData = ({
    file,
    revision,
    repo,
}: UseCodeInsightsDataInput): QueryResult<GetRelatedInsightsInlineResult> =>
    useQuery(RELATED_INSIGHTS_INLINE_QUERY, {
        variables: {
            input: {
                file,
                revision,
                repo,
            },
        },
    })
