import React from 'react'
import { Link } from '../../../../../shared/src/components/Link'
import { pluralize } from '../../../../../shared/src/util/strings'
import { AuthenticatedUser } from '../../../auth'
import { defaultExternalServices } from '../../../components/externalServices/externalServices'
import { ViewerCampaignsCodeHostsFields } from '../../../graphql-operations'

export interface CampaignSpecMissingCredentialsAlertProps {
    viewerCampaignsCodeHosts: ViewerCampaignsCodeHostsFields
    authenticatedUser: Pick<AuthenticatedUser, 'url'>
}

export const CampaignSpecMissingCredentialsAlert: React.FunctionComponent<CampaignSpecMissingCredentialsAlertProps> = ({
    viewerCampaignsCodeHosts,
    authenticatedUser,
}) => {
    if (viewerCampaignsCodeHosts.totalCount === 0) {
        return <></>
    }
    return (
        <div className="alert alert-warning">
            <h4 className="alert-heading">
                You don't have credentials configured for{' '}
                {pluralize('this code host', viewerCampaignsCodeHosts.totalCount, 'these code hosts')}
            </h4>
            <ul>
                {viewerCampaignsCodeHosts.nodes.map(node => (
                    <MissingCodeHost {...node} key={node.externalServiceKind + node.externalServiceURL} />
                ))}
            </ul>
            <p className="mb-0">
                Configure {pluralize('it', viewerCampaignsCodeHosts.totalCount, 'them')} in your{' '}
                <Link to={`${authenticatedUser.url}/settings/campaigns`} target="_blank" rel="noopener">
                    campaigns user settings
                </Link>{' '}
                to apply this spec.
            </p>
        </div>
    )
}

const MissingCodeHost: React.FunctionComponent<ViewerCampaignsCodeHostsFields['nodes'][0]> = ({
    externalServiceKind,
    externalServiceURL,
}) => {
    const Icon = defaultExternalServices[externalServiceKind].icon
    return (
        <li>
            <Icon className="icon-inline mr-2" />
            {externalServiceURL}
        </li>
    )
}
