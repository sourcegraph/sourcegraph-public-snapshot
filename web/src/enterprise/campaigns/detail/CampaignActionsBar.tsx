import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { CampaignsIcon } from '../icons'
import { Link } from '../../../../../shared/src/components/Link'
import { CloseDeleteCampaignPrompt } from './form/CloseDeleteCampaignPrompt'
import { CampaignUIMode } from './CampaignDetails'

interface Props {
    mode: CampaignUIMode

    campaign: Pick<GQL.ICampaign, 'name' | 'closedAt' | 'viewerCanAdminister'> & {
        changesets: {
            stats: Pick<GQL.ICampaign['changesets']['stats'], 'total' | 'closed' | 'merged'>
        }
    }

    onClose: (closeChangesets: boolean) => Promise<void>
}

export const CampaignActionsBar: React.FunctionComponent<Props> = ({ campaign, mode, onClose }) => {
    const campaignClosed = !!campaign.closedAt
    // TODO: New way to determine processing status
    const campaignProcessing = false
    const actionsDisabled = mode === 'deleting' || mode === 'closing' || campaignProcessing

    const percentComplete = (
        ((campaign.changesets.stats.closed + campaign.changesets.stats.merged) / campaign.changesets.stats.total) *
        100
    ).toFixed(0)

    return (
        <>
            <div className="mb-2">
                <span>
                    <Link to="/campaigns">Campaigns</Link>
                </span>
                <span className="text-muted d-inline-block mx-1">/</span>
                <span>{campaign.name}</span>
            </div>
            <div className="d-flex mb-2 position-relative">
                <div>
                    <h1 className="m-0">{campaign.name}</h1>
                    <h2 className="m-0">
                        <CampaignStateBadge isClosed={campaignClosed} />
                        <small className="text-muted">
                            {percentComplete}% complete . {campaign.changesets.stats.total} changesets total
                        </small>
                    </h2>
                </div>
                <span className="flex-grow-1 d-flex justify-content-end align-items-end">
                    {campaign.viewerCanAdminister && !campaignClosed && (
                        <CloseDeleteCampaignPrompt
                            disabled={actionsDisabled}
                            disabledTooltip="Cannot close while campaign is being created"
                            message={
                                <p>
                                    Close campaign <strong>{campaign.name}</strong>?
                                </p>
                            }
                            buttonText="Close"
                            onButtonClick={onClose}
                            buttonClassName="btn-secondary"
                        />
                    )}
                </span>
            </div>
        </>
    )
}

const CampaignStateBadge: React.FunctionComponent<{ isClosed: boolean }> = ({ isClosed }) => {
    if (isClosed) {
        return (
            <span className="badge badge-danger mr-2">
                <CampaignsIcon className="icon-inline campaign-actions-bar__campaign-icon" /> Closed
            </span>
        )
    }
    return (
        <span className="badge badge-success mr-2">
            <CampaignsIcon className="icon-inline campaign-actions-bar__campaign-icon" /> Open
        </span>
    )
}
