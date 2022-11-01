import { ApolloClient } from '@apollo/client'
import { groupBy } from 'lodash'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { isDefined } from '@sourcegraph/common'
import { fromObservableQuery } from '@sourcegraph/http-client'

import {
    InsightsDashboardCurrentUser,
    InsightsDashboardNode,
    InsightsDashboardsResult,
} from '../../../../../../../graphql-operations'
import { ALL_INSIGHTS_DASHBOARD } from '../../../../../constants'
import {
    InsightDashboard,
    InsightsDashboardOwner,
    InsightsDashboardOwnerType,
    InsightsDashboardType,
} from '../../../../types'
import { GET_INSIGHTS_DASHBOARDS_GQL } from '../../gql/GetInsightsDashboards'

export const getDashboards = (apolloClient: ApolloClient<unknown>, id?: string): Observable<InsightDashboard[]> =>
    fromObservableQuery(
        apolloClient.watchQuery<InsightsDashboardsResult>({
            query: GET_INSIGHTS_DASHBOARDS_GQL,
            variables: { id },
            fetchPolicy: 'cache-first',
        })
    ).pipe(
        map(result => {
            const {
                data: { insightsDashboards, currentUser },
            } = result

            return [
                ALL_INSIGHTS_DASHBOARD,
                ...makeDashboardTitleUnuque(insightsDashboards.nodes).map(
                    (dashboard): InsightDashboard => ({
                        id: dashboard.id,
                        type: InsightsDashboardType.Custom,
                        title: dashboard.title,
                        owners: deserializeDashboardsOwners(dashboard, currentUser),
                    })
                ),
            ]
        })
    )

function makeDashboardTitleUnuque(dashboards: InsightsDashboardNode[]): InsightsDashboardNode[] {
    const groupedByTitle = groupBy(dashboards, dashboard => dashboard.title)

    return Object.keys(groupedByTitle).flatMap(title => {
        if (groupedByTitle[title].length === 1) {
            return groupedByTitle[title]
        }

        return groupedByTitle[title].map((dashboard, index) => ({
            ...dashboard,
            title: `${dashboard.title} (${index + 1})`,
        }))
    })
}

function deserializeDashboardsOwners(
    dashboardNode: InsightsDashboardNode,
    userNode: InsightsDashboardCurrentUser | null
): InsightsDashboardOwner[] {
    if (!userNode) {
        return []
    }

    const {
        id: currentUserId,
        organizations: { nodes: organizations },
    } = userNode
    const {
        grants: { users: usersIds, organizations: organizationsIds, global },
    } = dashboardNode

    if (global) {
        return [
            {
                id: 'GLOBAL_INSTANCE_ID',
                type: InsightsDashboardOwnerType.Global,
                title: 'Global',
            },
        ]
    }

    const userOwners = usersIds
        .filter(userId => userId === currentUserId)
        .map<InsightsDashboardOwner>(userId => ({
            id: userId,
            type: InsightsDashboardOwnerType.Personal,
            title: 'Personal',
        }))

    const organizationOwners = organizationsIds
        .map<InsightsDashboardOwner | null>(orgId => {
            const organization = organizations.find(organization => orgId === organization.id)

            if (!organization) {
                return null
            }

            return {
                id: organization.id,
                type: InsightsDashboardOwnerType.Organization,
                title: organization.displayName ?? 'Unknown organization',
            }
        })
        .filter(isDefined)

    return [...userOwners, ...organizationOwners]
}
