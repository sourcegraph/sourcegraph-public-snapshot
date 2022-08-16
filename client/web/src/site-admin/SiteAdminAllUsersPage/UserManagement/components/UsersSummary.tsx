import React, { useMemo } from 'react'

import { useQuery } from '@sourcegraph/http-client'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { UsersManagementSummaryResult, UsersManagementSummaryVariables } from '../../../../graphql-operations'
import { ValueLegendList, ValueLegendListProps } from '../../../analytics/components/ValueLegendList'
import { USERS_MANAGEMENT_SUMMARY } from '../queries'

export const UsersSummary: React.FunctionComponent = () => {
    const { data, error, loading } = useQuery<UsersManagementSummaryResult, UsersManagementSummaryVariables>(
        USERS_MANAGEMENT_SUMMARY,
        {
            variables: {},
        }
    )

    const legends = useMemo(() => {
        if (!data) {
            return []
        }

        const legends: ValueLegendListProps['items'] = [
            {
                value: data.site.registeredUsers.totalCount,
                description: 'Registered Users',
                color: 'var(--purple)',
                position: 'left',
                tooltip: 'Total number of registered and not deleted users.',
            },
            {
                value: data.site.productSubscription.license?.userCount ?? 0,
                description: 'User licenses',
                color: 'var(--body-color)',
                position: 'right',
                tooltip: 'The number of user licenses your current account is provisioned for.',
            },
            {
                value: data.site.adminUsers?.totalCount ?? 0,
                description: 'Administrators',
                color: 'var(--body-color)',
                position: 'right',
                tooltip: 'The number of users with site admin permissions.',
            },
        ]

        return legends
    }, [data])

    if (error) {
        throw error
    }

    if (loading || !legends.length) {
        return <LoadingSpinner />
    }

    return <ValueLegendList className="mb-3" items={legends} />
}
