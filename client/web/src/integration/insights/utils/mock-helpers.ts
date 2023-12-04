import { testUserID } from '@sourcegraph/shared/src/testing/integration/graphQlResults'

import type {
    GetDashboardInsightsResult,
    GetInsightViewResult,
    InsightsDashboardNode,
    InsightViewNode,
} from '../../../graphql-operations'
import type { WebIntegrationTestContext } from '../../context'

import { type OverrideGraphQLExtensionsProps, overrideInsightsGraphQLApi } from './override-insights-graphql-api'

interface MakeOverridesOptions {
    testContext: WebIntegrationTestContext
    dashboardId: string
    insightMock: InsightViewNode
    insightViewMock?: GetInsightViewResult
    overrides?: OverrideGraphQLExtensionsProps['overrides']
}

/**
 * Helper function to remove some of the boiler plate when overriding gql calls.
 */
export function mockDashboardWithInsights(options: MakeOverridesOptions): void {
    const { testContext, dashboardId, insightMock, insightViewMock, overrides } = options

    const defaultOverrides: OverrideGraphQLExtensionsProps['overrides'] = {
        // Mock list of possible code insights dashboards on the dashboard page
        InsightsDashboards: () => ({
            currentUser: {
                __typename: 'User',
                id: testUserID,
                organizations: { nodes: [] },
            },
            insightsDashboards: {
                __typename: 'InsightsDashboardConnection',
                nodes: [createDashboard({ id: dashboardId })],
            },
        }),
        // Mock dashboard configuration (dashboard content) with one capture group insight configuration
        GetDashboardInsights: () =>
            createDashboardViewMock({
                id: dashboardId,
                insightsMocks: [insightMock],
            }),
    }

    if (insightViewMock) {
        defaultOverrides.GetInsightView = () => insightViewMock
    }

    overrideInsightsGraphQLApi({
        testContext,
        overrides: { ...defaultOverrides, ...overrides },
    })
}

interface DashboardMockOptions {
    id?: string
    title?: string
}

/**
 * Creates dashboard mock entity in order to mock insight dashboard configurations.
 * It's used for easier mocking dashboards list gql handler see InsightsDashboards entry point.
 */
function createDashboard(options: DashboardMockOptions): InsightsDashboardNode {
    const { id = '001_dashboard', title = 'Default dashboard' } = options

    return {
        __typename: 'InsightsDashboard',
        id,
        title,
        grants: {
            __typename: 'InsightsPermissionGrants',
            users: [testUserID],
            organizations: [],
            global: false,
        },
    }
}

interface DashboardViewMockOptions {
    id?: string
    insightsMocks?: InsightViewNode[]
}

/**
 * Creates mocks for the dashboard view entity, you may ask what difference between createDashboard
 * and this mock helper. Dashboard view mock contains insight configurations, dashboard mocks contain
 * only insight ids. (we should mock dashboard view only when this dashboard is opened on the page)
 */
function createDashboardViewMock(options: DashboardViewMockOptions): GetDashboardInsightsResult {
    const { id = '001_dashboard', insightsMocks = [] } = options

    return {
        insightsDashboards: {
            __typename: 'InsightsDashboardConnection',
            nodes: [
                {
                    __typename: 'InsightsDashboard',
                    id,
                    views: {
                        __typename: 'InsightViewConnection',
                        nodes: insightsMocks.map(mock => ({ ...mock, __typename: 'InsightView' })),
                    },
                },
            ],
        },
    }
}
