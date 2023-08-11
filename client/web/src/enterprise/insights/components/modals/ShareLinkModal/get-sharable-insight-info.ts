import { type ApolloError, gql, useQuery } from '@apollo/client'

import type {
    GetDashboardsThatHaveInsightResult,
    GetDashboardsThatHaveInsightVariables,
} from '../../../../../graphql-operations'
import { type CustomInsightDashboard, InsightsDashboardType } from '../../../core'
import { deserializeDashboardsOwners } from '../../../core/hooks/use-insight-dashboards'

export const GET_DASHBOARD_THAT_HAVE_INSIGHT_GQL = gql`
    query GetDashboardsThatHaveInsight($id: ID!) {
        insightViews(id: $id) {
            nodes {
                dashboards {
                    nodes {
                        ...InsightsDashboardNode
                    }
                }
            }
        }

        currentUser {
            ...InsightsDashboardCurrentUser
        }
    }

    fragment InsightsDashboardNode on InsightsDashboard {
        id
        title
        grants {
            users
            organizations
            global
        }
    }

    fragment InsightsDashboardCurrentUser on User {
        id
        organizations {
            nodes {
                id
                displayName
            }
        }
    }
`

interface Props {
    insightId: string
}

interface Result {
    dashboards: CustomInsightDashboard[] | null | undefined
    loading: boolean
    error: ApolloError | undefined
}

export function useDashboardThatHaveInsight(props: Props): Result {
    const { insightId } = props

    const { data, error, loading } = useQuery<
        GetDashboardsThatHaveInsightResult,
        GetDashboardsThatHaveInsightVariables
    >(GET_DASHBOARD_THAT_HAVE_INSIGHT_GQL, { variables: { id: insightId } })

    if (data) {
        const { insightViews, currentUser } = data
        const [insight] = insightViews.nodes
        const dashboards = insight?.dashboards?.nodes ?? []

        return {
            error,
            loading,
            dashboards: dashboards.map<CustomInsightDashboard>(dashboard => ({
                id: dashboard.id,
                type: InsightsDashboardType.Custom,
                title: dashboard.title,
                owners: deserializeDashboardsOwners(dashboard, currentUser),
            })),
        }
    }

    return { dashboards: undefined, error, loading }
}
