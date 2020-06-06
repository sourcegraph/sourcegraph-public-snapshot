import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { CampaignsIcon } from '../icons'
import { Link } from '../../../../../shared/src/components/Link'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { CloseDeleteCampaignPrompt } from './form/CloseDeleteCampaignPrompt'
import { CampaignUIMode } from './CampaignDetailArea'

interface Props {
    mode: CampaignUIMode

    campaign: Pick<GQL.ICampaign, 'name' | 'closedAt' | 'viewerCanAdminister'> & {
        status: Pick<GQL.ICampaign['status'], 'state'>
    }

    formID: string
}

export const CampaignBreadcrumbs: React.FunctionComponent<Props> = ({ campaign, mode, formID }) => {
    const showActionButtons = campaign.viewerCanAdminister

    const campaignClosed = campaign?.closedAt
    const campaignProcessing = campaign ? campaign.status.state === GQL.BackgroundProcessState.PROCESSING : false
    const actionsDisabled = mode === 'deleting' || mode === 'closing' || campaignProcessing

    let stateBadge: JSX.Element

    if (!campaign) {
        stateBadge = <CampaignsIcon className="icon-inline campaign-actions-bar__campaign-icon text-muted mr-2" />
    } else if (campaignClosed) {
        stateBadge = (
            <span className="badge badge-danger mr-2">
                <CampaignsIcon className="icon-inline campaign-actions-bar__campaign-icon" /> Closed
            </span>
        )
    } else {
        stateBadge = (
            <span className="badge badge-success mr-2">
                <CampaignsIcon className="icon-inline campaign-actions-bar__campaign-icon" /> Open
            </span>
        )
    }

    return (
        <div className="d-flex mb-2 position-relative">
            <h2 className="m-0">
                {stateBadge}
                <span>
                    <Link to="/campaigns">Campaigns</Link> <span className="badge badge-info">Beta</span>
                </span>
                <span className="text-muted d-inline-block mx-2">/</span>
                <span>{campaign.name}</span>
            </h2>
            <span className="flex-grow-1 d-flex justify-content-end align-items-center">actions TODO(sqs)</span>
        </div>
    )
}
