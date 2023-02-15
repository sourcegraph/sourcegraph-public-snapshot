import React from 'react'

import { useQuery } from '@sourcegraph/http-client'
import { ButtonLink, Badge } from '@sourcegraph/wildcard'

import { AccessRequestsCountResult, AccessRequestsCountVariables } from '../../graphql-operations'

import { ACCESS_REQUESTS_COUNT } from './queries'

export const AccessRequestsGlobalNavItem: React.FunctionComponent = () => {
    const { data } = useQuery<AccessRequestsCountResult, AccessRequestsCountVariables>(ACCESS_REQUESTS_COUNT, {
        fetchPolicy: 'network-only',
    })

    if (!data?.accessRequests.totalCount) {
        return null
    }

    return (
        <ButtonLink variant="danger" outline={true} size="sm" to="/site-admin/access-requests">
            Access requests
            <Badge variant="danger" pill={true} small={true} className="ml-1">
                {data?.accessRequests.totalCount}
            </Badge>
        </ButtonLink>
    )
}
