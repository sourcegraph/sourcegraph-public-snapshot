import type { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import { testUserID } from '@sourcegraph/shared/src/testing/integration/graphQlResults'

import type { WebGraphQlOperations } from '../../../graphql-operations'
import type { WebIntegrationTestContext } from '../../context'
import { commonWebGraphQlResults } from '../../graphQlResults'

export interface OverrideGraphQLExtensionsProps {
    testContext: WebIntegrationTestContext
    overrides?: Partial<WebGraphQlOperations & SharedGraphQlOperations>
}

/**
 * Test setup handler used for mocking common parts of API, extension insight API and
 * extension js bundle requests.
 *
 * @param props - Custom override for code insight APIs (gql, user setting, extensions)
 * @param [props.overrides] - GraphQL calls to override. Note: Overrides need all __typename fields included for Apollo cache to work.
 */
export function overrideInsightsGraphQLApi(props: OverrideGraphQLExtensionsProps): void {
    const { testContext, overrides = {} } = props

    // Mock temporary settings cause code insights beta modal UI relies on this handler to show/hide
    // modal UI on all code insights related pages.
    testContext.overrideInitialTemporarySettings({
        'insights.freeGaAccepted': true,
        'insights.wasMainPageOpen': true,
    })
    testContext.overrideGraphQL({
        ...commonWebGraphQlResults,
        // Mock insight config query
        GetInsights: () => ({
            __typename: 'Query',
            insightViews: {
                __typename: 'InsightViewConnection',
                nodes: [],
            },
        }),
        IsCodeInsightsLicensed: () => ({ __typename: 'Query', enterpriseLicenseHasFeature: true }),
        InsightsDashboards: () => ({
            currentUser: {
                __typename: 'User',
                id: testUserID,
                organizations: {
                    nodes: [
                        {
                            id: 'Org_test_id',
                            name: 'test organization',
                            displayName: 'Test organization',
                        },
                    ],
                },
            },
            insightsDashboards: {
                __typename: 'InsightsDashboardConnection',
                nodes: [],
            },
        }),

        GetAllInsightConfigurations: () => ({
            __typename: 'Query',
            insightViews: {
                __typename: 'InsightViewConnection',
                nodes: [],
                pageInfo: { __typename: 'PageInfo', endCursor: null, hasNextPage: false },
                totalCount: 0,
            },
        }),

        CurrentAuthState: () => ({
            currentUser: {
                __typename: 'User',
                id: testUserID,
                databaseID: 1,
                username: 'test',
                avatarURL: null,
                email: 'vova@sourcegraph.com',
                displayName: null,
                siteAdmin: true,
                tosAccepted: true,
                url: '/users/test',
                settingsURL: '/users/test/settings',
                hasVerifiedEmail: true,
                organizations: {
                    nodes: [
                        {
                            __typename: 'Org',
                            name: 'test organization',
                            displayName: 'Test organization',
                            id: 'Org_test_id',
                            settingsURL: '/organizations/test_organization/settings',
                            url: '/organizations/test_organization/settings',
                        },
                    ],
                },
                session: { canSignOut: true },
                viewerCanAdminister: true,
                emails: [],
                latestSettings: null,
                permissions: { nodes: [] },
            },
        }),
        ...overrides,
    })
}
