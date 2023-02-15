import React from 'react'

import { useQuery } from '@sourcegraph/http-client'
import { ButtonLink, Badge } from '@sourcegraph/wildcard'

import { AccessRequestsCountResult, AccessRequestsCountVariables } from '../../graphql-operations'
import { SourcegraphContext } from '../../jscontext'
import { checkIsRequestAccessEnabled } from '../../util/checkIsRequestAccessEnabled'

import { ACCESS_REQUESTS_COUNT } from './queries'

interface AccessRequestsGlobalNavItemProps {
    context: Pick<SourcegraphContext, 'allowSignup' | 'experimentalFeatures'>
    isSourcegraphDotCom: boolean
}

export const AccessRequestsGlobalNavItem: React.FunctionComponent<AccessRequestsGlobalNavItemProps> = props => {
    const isRequestAccessEnabled = checkIsRequestAccessEnabled(
        props.isSourcegraphDotCom,
        props.context.allowSignup,
        props.context.experimentalFeatures.requestAccess
    )
    const { data } = useQuery<AccessRequestsCountResult, AccessRequestsCountVariables>(ACCESS_REQUESTS_COUNT, {
        fetchPolicy: 'network-only',
        skip: !isRequestAccessEnabled,
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
