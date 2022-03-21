import React, { useState } from 'react'

import { storiesOf } from '@storybook/react'
import { of } from 'rxjs'

import { WebStory } from '../../../../../../../components/WebStory'
import { CodeInsightsBackendStoryMock } from '../../../../../CodeInsightsBackendStoryMock'
import { AccessibleInsightInfo } from '../../../../../core/backend/code-insights-backend-types'
import { InsightsDashboardType, InsightsDashboardScope, CustomInsightDashboard } from '../../../../../core/types'

import { AddInsightModal } from './AddInsightModal'

const { add } = storiesOf('web/insights/AddInsightModal', module)
    .addDecorator(story => <WebStory>{() => story()}</WebStory>)
    .addParameters({
        chromatic: {
            viewports: [576, 1440],
        },
    })

const dashboard: CustomInsightDashboard = {
    type: InsightsDashboardType.Custom,
    scope: InsightsDashboardScope.Personal,
    id: '001',
    title: 'Test dashboard',
    insightIds: [],
    owner: {
        id: 'user_test_id',
        name: 'Emir Kusturica',
    },
}

const mockInsights: AccessibleInsightInfo[] = [
    {
        id: 'searchInsights.insight.personalGraphQLTypesMigration',
        title: '[Personal] Migration to new GraphQL TS types',
    },
    {
        id: 'searchInsights.insight.testOrg1graphQLTypesMigration',
        title:
            '[Test ORG 1] Migration to new GraphQL TS types [Test ORG 1] Migration to new GraphQL TS types [Test ORG 1] Migration to new GraphQL TS types',
    },
    {
        id: 'searchInsights.insight.testOrg1graphQLTypesMigration1',
        title: '[Test ORG 1] Migration to new GraphQL TS types',
    },
    {
        id: 'searchInsights.insight.testOrg1graphQLTypesMigration2',
        title: '[Test ORG 1] Migration to new GraphQL TS types',
    },
    {
        id: 'searchInsights.insight.testOrg2graphQLTypesMigration',
        title: '[Test ORG 2] Migration to new GraphQL TS types',
    },
]

const codeInsightsBackend = {
    getReachableInsights: () => of(mockInsights),
    getDashboardSubjects: () => of([]),
}

add('AddInsightModal', () => {
    const [open, setOpen] = useState<boolean>(true)

    return (
        <CodeInsightsBackendStoryMock mocks={codeInsightsBackend}>
            {open && <AddInsightModal dashboard={dashboard} onClose={() => setOpen(false)} />}
        </CodeInsightsBackendStoryMock>
    )
})
