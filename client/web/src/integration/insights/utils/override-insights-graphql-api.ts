import { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import { testUserID } from '@sourcegraph/shared/src/testing/integration/graphQlResults'

import {
    BulkSearchCommits,
    BulkSearchFields,
    BulkSearchRepositories,
    WebGraphQlOperations,
} from '../../../graphql-operations'
import { WebIntegrationTestContext } from '../../context'
import { commonWebGraphQlResults } from '../../graphQlResults'

/**
 * Some insight creation UI gql api requests do not have
 * generated types due their dynamic nature. Because of that we
 * must write these api call types below manually for testing purposes.
 */
interface CustomInsightsOperations {
    /** API handler used for repositories field async validation. */
    BulkRepositoriesSearch: () => Record<string, BulkSearchRepositories>

    /** Internal API handler for fetching commits data for live preview chart. */
    BulkSearchCommits: () => Record<string, BulkSearchCommits>

    /**
     * Internal API handler for fetching actual data according to search commits
     * for live preview chart.
     */
    BulkSearch: () => Record<string, BulkSearchFields>
}

interface OverrideGraphQLExtensionsProps {
    testContext: WebIntegrationTestContext
    overrides?: Partial<WebGraphQlOperations & SharedGraphQlOperations & CustomInsightsOperations>
}

/**
 * Test setup handler used for mocking common parts of API, extension insight API and
 * extension js bundle requests.
 *
 * @param props - Custom override for code insight APIs (gql, user setting, extensions)
 */
export function overrideInsightsGraphQLApi(props: OverrideGraphQLExtensionsProps): void {
    const { testContext, overrides = {} } = props

    testContext.overrideGraphQL({
        ...commonWebGraphQlResults,
        // Mock temporary settings cause code insights beta modal UI relies on this handler to show/hide
        // modal UI on all code insights related pages.
        GetTemporarySettings: () => ({
            temporarySettings: {
                __typename: 'TemporarySettings',
                contents: JSON.stringify({
                    'user.daysActiveCount': 1,
                    'user.lastDayActive': new Date().toDateString(),
                    'search.usedNonGlobalContext': true,
                    'insights.freeGaAccepted': true,
                    'insights.wasMainPageOpen': true,
                }),
            },
        }),
        // Mock insight config query
        GetInsights: () => ({
            __typename: 'Query',
            insightViews: {
                __typename: 'InsightViewConnection',
                nodes: [],
            },
        }),
        HasAvailableCodeInsight: () => ({
            insightViews: {
                __typename: 'InsightViewConnection',
                nodes: [{ id: '001', isFrozen: false }],
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
                tags: [],
                tosAccepted: true,
                url: '/users/test',
                settingsURL: '/users/test/settings',
                organizations: {
                    nodes: [
                        {
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
                searchable: true,
            },
        }),
        ...overrides,
    })
}
