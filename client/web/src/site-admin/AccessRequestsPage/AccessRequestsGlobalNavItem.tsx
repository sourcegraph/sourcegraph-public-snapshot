import React from 'react'

import { useQuery } from '@sourcegraph/http-client'
import { ButtonLink, Badge } from '@sourcegraph/wildcard'

import { AccessRequestsCountResult, AccessRequestsCountVariables } from '../../graphql-operations'
import { SourcegraphContext } from '../../jscontext'
import { checkIsRequestAccessAllowed } from '../../util/checkIsRequestAccessAllowed'

import { ACCESS_REQUESTS_COUNT } from './queries'

interface AccessRequestsGlobalNavItemProps {
    context: Pick<SourcegraphContext, 'allowSignup' | 'experimentalFeatures'>
    isSourcegraphDotCom: boolean
}

/**
 * A link to the access requests page that shows a badge with the number of pending requests.
 * Does not render anything if request access is not allowed or there are no pending requests.
 */
export const AccessRequestsGlobalNavItem: React.FunctionComponent<AccessRequestsGlobalNavItemProps> = props => {
    const isRequestAccessAllowed = checkIsRequestAccessAllowed(
        props.isSourcegraphDotCom,
        props.context.allowSignup,
        props.context.experimentalFeatures['accessRequests.enabled']
    )

    const { data } = useQuery<AccessRequestsCountResult, AccessRequestsCountVariables>(ACCESS_REQUESTS_COUNT, {
        fetchPolicy: 'network-only',
        skip: !isRequestAccessAllowed,
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
