import { testUserID } from '@sourcegraph/shared/src/testing/integration/graphQlResults'

import { InsightsDashboardsResult } from '../../../graphql-operations'

export const GET_DASHBOARD_INSIGHTS = {
    insightsDashboards: {
        nodes: [
            {
                id: 'EMPTY_DASHBOARD',
                views: { nodes: [] },
            },
        ],
    },
}

export const INSIGHTS_DASHBOARDS: InsightsDashboardsResult = {
    currentUser: {
        __typename: 'User',
        id: testUserID,
        organizations: {
            __typename: 'OrgConnection',
            nodes: [
                {
                    __typename: 'Org',
                    id: 'Org_test_id',
                    displayName: 'Test organization OVERRIDDEN',
                },
            ],
        },
    },
    insightsDashboards: {
        __typename: 'InsightsDashboardConnection',
        nodes: [
            {
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
            },
        ],
    },
}
