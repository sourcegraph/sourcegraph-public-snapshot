import { type ApolloClient, gql } from '@apollo/client'
import { from, type Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import type {
    CreateDashboardResult,
    CreateDashboardVariables,
    InsightsPermissionGrantsInput,
} from '../../../../../../../graphql-operations'
import { type InsightsDashboardOwner, InsightsDashboardOwnerType } from '../../../../types'
import type { DashboardCreateInput, DashboardCreateResult } from '../../../code-insights-backend-types'

const CREATE_DASHBOARD_MUTATION_GQL = gql`
    mutation CreateDashboard($input: CreateInsightsDashboardInput!) {
        createInsightsDashboard(input: $input) {
            dashboard {
                id
                title
                views {
                    nodes {
                        id
                    }
                }
                grants {
                    users
                    organizations
                    global
                }
            }
        }
    }
`

const CACHE_DASHBOARD_UPDATE_FRAGMENT = gql`
    fragment NewDashboard on InsightsDashboard {
        id
        title
        views {
            nodes {
                id
            }
        }
        grants {
            users
            organizations
            global
        }
    }
`

export const createDashboard = (
    apolloClient: ApolloClient<unknown>,
    input: DashboardCreateInput
): Observable<DashboardCreateResult> => {
    const { name, owners } = input
    return from(
        apolloClient.mutate<CreateDashboardResult, CreateDashboardVariables>({
            mutation: CREATE_DASHBOARD_MUTATION_GQL,
            variables: { input: { title: name, grants: serializeDashboardOwners(owners) } },
            update(cache, result) {
                const { data } = result

                if (!data) {
                    return
                }

                cache.modify({
                    fields: {
                        insightsDashboards(dashboards) {
                            const newDashboardsReference = cache.writeFragment({
                                data: data.createInsightsDashboard.dashboard,
                                fragment: CACHE_DASHBOARD_UPDATE_FRAGMENT,
                            })

                            return { nodes: [...(dashboards.nodes ?? []), newDashboardsReference] }
                        },
                    },
                })
            },
        })
    ).pipe(
        map(result => ({
            id: result.data?.createInsightsDashboard.dashboard.id ?? 'unknown',
        }))
    )
}

export function serializeDashboardOwners(owners: InsightsDashboardOwner[]): InsightsPermissionGrantsInput {
    const hasGlobalOwner = owners.some(owner => owner.type === InsightsDashboardOwnerType.Global)

    if (hasGlobalOwner) {
        return { users: [], organizations: [], global: true }
    }

    return {
        users: owners.filter(owner => owner.type === InsightsDashboardOwnerType.Personal).map(owner => owner.id),
        organizations: owners
            .filter(owner => owner.type === InsightsDashboardOwnerType.Organization)
            .map(owner => owner.id),
        global: false,
    }
}
