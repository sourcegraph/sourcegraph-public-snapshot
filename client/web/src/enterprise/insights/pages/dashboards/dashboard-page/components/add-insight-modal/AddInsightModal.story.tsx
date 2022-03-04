/* eslint-disable @typescript-eslint/no-floating-promises */
import { storiesOf } from '@storybook/react'
import React, { useState } from 'react'
import { of } from 'rxjs'

import { WebStory } from '../../../../../../../components/WebStory'
import { CodeInsightsBackendStoryMock } from '../../../../../CodeInsightsBackendStoryMock'
import { ReachableInsight } from '../../../../../core/backend/code-insights-backend-types'
import {
    InsightsDashboardType,
    InsightsDashboardScope,
    CustomInsightDashboard,
    InsightExecutionType,
    InsightType,
} from '../../../../../core/types'

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

const mockInsights: ReachableInsight[] = [
    {
        id: 'searchInsights.insight.personalGraphQLTypesMigration',
        title: '[Personal] Migration to new GraphQL TS types',
        repositories: ['github.com/sourcegraph/sourcegraph'],
        series: [],
        step: { weeks: 6 },
        owner: {
            id: 'user_test_id',
            name: 'test',
        },
    },
    {
        id: 'searchInsights.insight.testOrg1graphQLTypesMigration',
        title:
            '[Test ORG 1] Migration to new GraphQL TS types [Test ORG 1] Migration to new GraphQL TS types [Test ORG 1] Migration to new GraphQL TS types',
        repositories: ['github.com/sourcegraph/sourcegraph'],
        series: [],
        step: { weeks: 6 },
        owner: {
            id: 'test_org_1_id',
            name: 'Test organization 1 Test organization 1 Test organization 1',
        },
    },
    {
        id: 'searchInsights.insight.testOrg1graphQLTypesMigration1',
        title: '[Test ORG 1] Migration to new GraphQL TS types',
        repositories: ['github.com/sourcegraph/sourcegraph'],
        series: [],
        step: { weeks: 6 },
        owner: {
            id: 'test_org_1_id',
            name: 'Test organization 1 Test organization 1 Test organization 1',
        },
    },
    {
        id: 'searchInsights.insight.testOrg1graphQLTypesMigration2',
        title: '[Test ORG 1] Migration to new GraphQL TS types',
        repositories: ['github.com/sourcegraph/sourcegraph'],
        series: [],
        step: { weeks: 6 },
        owner: {
            id: 'test_org_1_id',
            name: 'Test organization 1 Test organization 1 Test organization 1',
        },
    },
    {
        id: 'searchInsights.insight.testOrg2graphQLTypesMigration',
        title: '[Test ORG 2] Migration to new GraphQL TS types',
        repositories: ['github.com/sourcegraph/sourcegraph'],
        series: [],
        step: { weeks: 6 },
        owner: {
            id: 'test_org_2_id',
            name: 'Test organization 2',
        },
    },
].map(insight => ({
    ...insight,
    dashboardReferenceCount: 0,
    otherThreshold: 0,
    query: '',
    type: InsightExecutionType.Backend,
    viewType: InsightType.SearchBased,
    visibility: 'global',
}))

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
