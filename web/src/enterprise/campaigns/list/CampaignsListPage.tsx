import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React from 'react'
import { Link } from 'react-router-dom'
import { ExtensionsControllerNotificationProps } from '../../../../../shared/src/extensions/controller'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { pluralize } from '../../../../../shared/src/util/strings'
import { NamespaceAreaContext } from '../../../namespaces/NamespaceArea'
import { CampaignListItem } from './CampaignListItem'
import { useCampaignsDefinedInNamespace } from './useCampaignsDefinedInNamespace'

const LOADING: 'loading' = 'loading'

interface Props extends Pick<NamespaceAreaContext, 'namespace'>, ExtensionsControllerNotificationProps {
    newCampaignURL: string
}

/**
 * Lists a namespace's campaigns.
 */
export const CampaignsListPage: React.FunctionComponent<Props> = ({ namespace, newCampaignURL, ...props }) => {
    const [campaignsOrError] = useCampaignsDefinedInNamespace(namespace)
    return (
        <div className="campaigns-list-page">
            <Link to={newCampaignURL} className="btn btn-primary mb-3">
                New campaign
            </Link>
            {campaignsOrError === LOADING ? (
                <LoadingSpinner className="icon-inline mt-3" />
            ) : isErrorLike(campaignsOrError) ? (
                <div className="alert alert-danger mt-3">{campaignsOrError.message}</div>
            ) : (
                <div className="card">
                    <div className="card-header">
                        <span className="text-muted">
                            {campaignsOrError.totalCount} {pluralize('campaign', campaignsOrError.totalCount)}
                        </span>
                    </div>
                    {campaignsOrError.nodes.length > 0 ? (
                        <ul className="list-group list-group-flush">
                            {campaignsOrError.nodes.map(campaign => (
                                <li key={campaign.id} className="list-group-item">
                                    <CampaignListItem {...props} campaign={campaign} />
                                </li>
                            ))}
                        </ul>
                    ) : (
                        <div className="p-2 text-muted">No campaigns yet.</div>
                    )}
                </div>
            )}
        </div>
    )
}
