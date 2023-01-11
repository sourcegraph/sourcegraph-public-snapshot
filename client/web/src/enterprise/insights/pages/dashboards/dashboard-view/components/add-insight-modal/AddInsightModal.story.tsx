import { useState } from 'react'

import { MockedResponse } from '@apollo/client/testing/core'
import { DecoratorFn, Story, Meta } from '@storybook/react'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo/mockedTestProvider'

import { WebStory } from '../../../../../../../components/WebStory'
import { GetDashboardAccessibleInsightsResult } from '../../../../../../../graphql-operations'
import { CustomInsightDashboard, InsightsDashboardOwnerType, InsightsDashboardType } from '../../../../../core'

import { AddInsightModal } from './AddInsightModal'
import { GET_ACCESSIBLE_INSIGHTS_LIST } from './query'

const decorator: DecoratorFn = story => <WebStory>{() => story()}</WebStory>

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

const mockInsights: MockedResponse<GetDashboardAccessibleInsightsResult> = {
    request: {
        query: getDocumentNode(GET_ACCESSIBLE_INSIGHTS_LIST),
        variables: { id: '001' },
    },
    result: {
        data: {
            dashboardInsightsIds: { nodes: [{ views: { nodes: [] } }] },
            accessibleInsights: {
                __typename: 'InsightViewConnection',
                nodes: [
                    {
                        __typename: 'InsightView',
                        id: 'searchInsights.insight.personalGraphQLTypesMigration',
                        presentation: {
                            __typename: 'LineChartInsightViewPresentation',
                            title: '[Personal] Migration to new GraphQL TS types',
                        },
                    },
                    {
                        __typename: 'InsightView',
                        id: 'searchInsights.insight.testOrg1graphQLTypesMigration',
                        presentation: {
                            __typename: 'LineChartInsightViewPresentation',
                            title: '[Test ORG 1] Migration to new GraphQL TS types [Test ORG 1] Migration to new GraphQL TS types [Test ORG 1] Migration to new GraphQL TS types',
                        },
                    },
                    {
                        __typename: 'InsightView',
                        id: 'searchInsights.insight.testOrg1graphQLTypesMigration1',
                        presentation: {
                            __typename: 'LineChartInsightViewPresentation',
                            title: '[Test ORG 1] Migration to new GraphQL TS types',
                        },
                    },
                    {
                        __typename: 'InsightView',
                        id: 'searchInsights.insight.testOrg1graphQLTypesMigration2',
                        presentation: {
                            __typename: 'LineChartInsightViewPresentation',
                            title: '[Test ORG 1] Migration to new GraphQL TS types',
                        },
                    },
                    {
                        __typename: 'InsightView',
                        id: 'searchInsights.insight.testOrg2graphQLTypesMigration',
                        presentation: {
                            __typename: 'LineChartInsightViewPresentation',
                            title: '[Test ORG 2] Migration to new GraphQL TS types',
                        },
                    },
                ],
            },
        },
    },
}

export const AddInsightModalStory: Story = () => {
    const [open, setOpen] = useState<boolean>(true)

    return (
        <MockedTestProvider mocks={[mockInsights]}>
            {open && <AddInsightModal dashboard={dashboard} onClose={() => setOpen(false)} />}
        </MockedTestProvider>
    )
}

AddInsightModalStory.storyName = 'AddInsightModal'
