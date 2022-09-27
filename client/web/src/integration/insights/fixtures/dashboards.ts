import { testUserID } from '@sourcegraph/shared/src/testing/integration/graphQlResults'

import {
    GetDashboardInsightsResult,
    InsightsDashboardNode,
    InsightsDashboardsResult,
} from '../../../graphql-operations'

export const EMPTY_DASHBOARD: InsightsDashboardNode = {
    __typename: 'InsightsDashboard',
    id: 'EMPTY_DASHBOARD',
    title: 'Empty Dashboard',
    views: {
        __typename: 'InsightViewConnection',
        nodes: [],
    },
    grants: {
        __typename: 'InsightsPermissionGrants',
        users: [testUserID],
        organizations: [],
        global: false,
    },
}

export const GET_DASHBOARD_INSIGHTS: GetDashboardInsightsResult = {
    insightsDashboards: {
        nodes: [
            {
                __typename: 'InsightsDashboard',
                id: EMPTY_DASHBOARD.id,
                views: { nodes: [] },
            },
        ],
    },
}

export const INSIGHTS_DASHBOARDS: InsightsDashboardsResult = {
    currentUser: {
        __typename: 'User',
        id: testUserID,
        organizations: { nodes: [] },
    },
    insightsDashboards: {
        __typename: 'InsightsDashboardConnection',
        nodes: [EMPTY_DASHBOARD],
    },
}
