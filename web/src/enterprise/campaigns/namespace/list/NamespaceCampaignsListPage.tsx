import React from 'react'
import { Link } from 'react-router-dom'
import { ExtensionsControllerNotificationProps } from '../../../../../../shared/src/extensions/controller'
import { NamespaceAreaContext } from '../../../../namespaces/NamespaceArea'
import { CampaignsList } from '../../list/CampaignsList'
import { useCampaigns } from '../../list/useCampaignsDefinedInNamespace'

interface Props extends Pick<NamespaceAreaContext, 'namespace'>, ExtensionsControllerNotificationProps {
    newCampaignURL: string | null
}

/**
 * Lists a namespace's campaigns.
 */
export const NamespaceCampaignsListPage: React.FunctionComponent<Props> = ({ newCampaignURL, namespace, ...props }) => {
    const campaigns = useCampaigns(namespace)
    return (
        <div className="namespace-campaigns-list-page">
            {newCampaignURL && (
                <Link to={newCampaignURL} className="btn btn-primary mb-3">
                    New campaign
                </Link>
            )}
            <CampaignsList {...props} campaigns={campaigns} />
        </div>
    )
}
