import type { ApolloClient } from '@apollo/client'
import type { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { fromObservableQuery } from '@sourcegraph/http-client'

import type { InsightSubjectsResult } from '../../../../../../graphql-operations'
import { type InsightsDashboardOwner, InsightsDashboardOwnerType } from '../../../types'
import { GET_INSIGHTS_DASHBOARD_OWNERS_GQL } from '../gql/GetInsightSubjects'

export const getDashboardOwners = (apolloClient: ApolloClient<unknown>): Observable<InsightsDashboardOwner[]> =>
    fromObservableQuery(
        apolloClient.watchQuery<InsightSubjectsResult>({
            query: GET_INSIGHTS_DASHBOARD_OWNERS_GQL,
        })
    ).pipe(
        map(({ data }) => {
            const { currentUser, site } = data

            const globalOwner: InsightsDashboardOwner = {
                id: site.id,
                type: InsightsDashboardOwnerType.Global,
                title: 'Global',
            }

            if (!currentUser) {
                return [globalOwner]
            }

            const userOwner: InsightsDashboardOwner = {
                id: currentUser.id,
                type: InsightsDashboardOwnerType.Personal,
                title: 'Personal',
            }

            const organizationOwners = currentUser.organizations.nodes.map(organization => ({
                id: organization.id,
                type: InsightsDashboardOwnerType.Organization,
                title: organization.displayName ?? organization.name,
            }))

            return [userOwner, ...organizationOwners, globalOwner]
        })
    )
