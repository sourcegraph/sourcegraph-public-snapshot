import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React from 'react'
import { ExtensionsControllerNotificationProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { pluralize } from '../../../../../shared/src/util/strings'
import { CampaignListItem } from './CampaignListItem'

const LOADING: 'loading' = 'loading'

interface Props extends ExtensionsControllerNotificationProps {
    campaigns: typeof LOADING | GQL.ICampaignConnection | ErrorLike
}

/**
 * Lists campaigns.
 */
export const CampaignList: React.FunctionComponent<Props> = ({ campaigns, ...props }) => (
    <div className="campaign-list">
        {campaigns === LOADING ? (
            <LoadingSpinner className="icon-inline mt-3" />
        ) : isErrorLike(campaigns) ? (
            <div className="alert alert-danger mt-3">{campaigns.message}</div>
        ) : (
            <div className="card">
                <div className="card-header">
                    <span className="text-muted">
                        {campaigns.totalCount} {pluralize('campaign', campaigns.totalCount)}
                    </span>
                </div>
                {campaigns.nodes.length > 0 ? (
                    <ul className="list-group list-group-flush">
                        {campaigns.nodes.map(campaign => (
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
