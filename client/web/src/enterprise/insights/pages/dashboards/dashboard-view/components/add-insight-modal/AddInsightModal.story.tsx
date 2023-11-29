import { useState } from 'react'

import type { MockedResponse } from '@apollo/client/testing/core'
import type { Decorator, StoryFn, Meta } from '@storybook/react'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo/mockedTestProvider'

import { WebStory } from '../../../../../../../components/WebStory'
import type {
    FindInsightsBySearchTermResult,
    FindInsightsBySearchTermVariables,
} from '../../../../../../../graphql-operations'
import { type CustomInsightDashboard, InsightsDashboardOwnerType, InsightsDashboardType } from '../../../../../core'

import { AddInsightModal } from './AddInsightModal'
import { GET_INSIGHTS_BY_SEARCH_TERM } from './query'

const decorator: Decorator = story => <WebStory>{() => story()}</WebStory>

const config: Meta = {
    title: 'web/insights/AddInsightModal',
    decorators: [decorator],
}

export default config

const dashboard: CustomInsightDashboard = {
    type: InsightsDashboardType.Custom,
    id: '001',
    title: 'Test dashboard',
    owners: [{ id: '001', title: 'Hieronymus Bosch', type: InsightsDashboardOwnerType.Personal }],
}

const mockInsights: MockedResponse<FindInsightsBySearchTermResult> = {
    request: {
        query: getDocumentNode(GET_INSIGHTS_BY_SEARCH_TERM),
        variables: { search: '', first: 20, after: null } as FindInsightsBySearchTermVariables,
    },
    result: {
        data: {
            insightViews: {
                __typename: 'InsightViewConnection',
                totalCount: 5,
                pageInfo: {
                    __typename: 'PageInfo',
                    hasNextPage: false,
                    endCursor: null,
                },
                nodes: [
                    {
                        __typename: 'InsightView',
                        id: 'searchInsights.insight.personalGraphQLTypesMigration',
                        presentation: {
                            __typename: 'LineChartInsightViewPresentation',
                            title: '[Personal] Migration to new GraphQL TS types',
                        },
                        dataSeriesDefinitions: [],
                    },
                    {
                        __typename: 'InsightView',
                        id: 'searchInsights.insight.testOrg1graphQLTypesMigration',
                        presentation: {
                            __typename: 'LineChartInsightViewPresentation',
                            title: '[Test ORG 1] Migration to new GraphQL TS types [Test ORG 1] Migration to new GraphQL TS types [Test ORG 1] Migration to new GraphQL TS types',
                        },
                        dataSeriesDefinitions: [],
                    },
                    {
                        __typename: 'InsightView',
                        id: 'searchInsights.insight.testOrg1graphQLTypesMigration1',
                        presentation: {
                            __typename: 'LineChartInsightViewPresentation',
                            title: '[Test ORG 1] Migration to new GraphQL TS types',
                        },
                        dataSeriesDefinitions: [],
                    },
                    {
                        __typename: 'InsightView',
                        id: 'searchInsights.insight.testOrg1graphQLTypesMigration2',
                        presentation: {
                            __typename: 'LineChartInsightViewPresentation',
                            title: '[Test ORG 1] Migration to new GraphQL TS types',
                        },
                        dataSeriesDefinitions: [],
                    },
                    {
                        __typename: 'InsightView',
                        id: 'searchInsights.insight.testOrg2graphQLTypesMigration',
                        presentation: {
                            __typename: 'LineChartInsightViewPresentation',
                            title: '[Test ORG 2] Migration to new GraphQL TS types',
                        },
                        dataSeriesDefinitions: [],
                    },
                ],
            },
        },
    },
}

export const AddInsightModalStory: StoryFn = () => {
    const [open, setOpen] = useState<boolean>(true)

    return (
        <MockedTestProvider mocks={[mockInsights]}>
            {open && <AddInsightModal dashboard={dashboard} onClose={() => setOpen(false)} />}
        </MockedTestProvider>
    )
}

AddInsightModalStory.storyName = 'AddInsightModal'
