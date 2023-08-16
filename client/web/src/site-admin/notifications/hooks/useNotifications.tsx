import { gql, useQuery } from '@sourcegraph/http-client'

import { NotificationsResult, NotificationsVariables } from '../../../graphql-operations'

const QUERY = gql`
    query Notifications {
        site {
            alerts {
                type
                group
                message
                isDismissibleWithKey
            }
        }
    }
`

interface UseNotificationsReturnType {
    data: NotificationsResult['site']['alerts']
    loading: boolean
    error?: any
}

export function useNotifications(): UseNotificationsReturnType {
    const { data, loading, error } = useQuery<NotificationsResult, NotificationsVariables>(QUERY, {})

    return {
        data: data?.site.alerts || [],
        loading,
        error,
    }
}
