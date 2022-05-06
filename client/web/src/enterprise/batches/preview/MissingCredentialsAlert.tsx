import React from 'react'

import { pluralize } from '@sourcegraph/common'
import { Alert, Link } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../auth'
import { ViewerBatchChangesCodeHostsFields } from '../../../graphql-operations'
import { CodeHost } from '../CodeHost'

export interface MissingCredentialsAlertProps {
    viewerBatchChangesCodeHosts: ViewerBatchChangesCodeHostsFields
    authenticatedUser: Pick<AuthenticatedUser, 'url'>
}

export const MissingCredentialsAlert: React.FunctionComponent<
    React.PropsWithChildren<MissingCredentialsAlertProps>
> = ({ viewerBatchChangesCodeHosts, authenticatedUser }) => {
    if (viewerBatchChangesCodeHosts.totalCount === 0) {
        return <></>
    }
    return (
        <Alert variant="warning">
            <p>
                <strong>
                    You don't have credentials configured for{' '}
                    {pluralize('this code host', viewerBatchChangesCodeHosts.totalCount, 'these code hosts')}
                </strong>
            </p>
            <ul>
                {viewerBatchChangesCodeHosts.nodes.map(node => (
                    <CodeHost {...node} key={node.externalServiceKind + node.externalServiceURL} />
                ))}
            </ul>
            <p className="mb-0">
                Credentials are required to publish changesets on code hosts. Configure them in your{' '}
                <Link to={`${authenticatedUser.url}/settings/batch-changes`} target="_blank" rel="noopener">
                    batch changes user settings
                </Link>{' '}
                to apply this spec.
            </p>
        </Alert>
    )
}
