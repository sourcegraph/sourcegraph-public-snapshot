import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { CampaignsIcon } from '../icons'
import { Link } from '../../../../../shared/src/components/Link'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { CloseDeleteCampaignPrompt } from './form/CloseDeleteCampaignPrompt'
import { CampaignUIMode } from './CampaignDetails'

interface Props {
    mode: CampaignUIMode
    previewingCampaignPlan: boolean

    campaign?: Pick<GQL.ICampaign, 'name' | 'closedAt' | 'viewerCanAdminister' | 'publishedAt'> & {
        changesets: Pick<GQL.ICampaign['changesets'], 'totalCount'> & {
            nodes: Pick<GQL.IExternalChangeset, 'state'>[]
        }
        status: Pick<GQL.ICampaign['status'], 'state'>
    }

    onClose: (closeChangesets: boolean) => Promise<void>
    onDelete: (closeChangesets: boolean) => Promise<void>
    onEdit: React.MouseEventHandler
}

export const CampaignActionsBar: React.FunctionComponent<Props> = ({
    campaign,
    previewingCampaignPlan,
    mode,
    onClose,
    onDelete,
    onEdit,
}) => {
    const showActionButtons = campaign && !previewingCampaignPlan && campaign.viewerCanAdminister
    const showSpinner = mode === 'saving' || mode === 'deleting' || mode === 'closing'
    const editingCampaign = mode === 'editing' || mode === 'saving'

    const campaignProcessing = campaign ? campaign.status.state === GQL.BackgroundProcessState.PROCESSING : false
    const actionsDisabled = mode === 'deleting' || mode === 'closing' || campaignProcessing

    const openChangesetsCount =
        campaign?.changesets.nodes.filter(changeset => changeset.state === GQL.ChangesetState.OPEN).length ?? 0

    const newCampaignHeader = previewingCampaignPlan ? 'New campaign' : 'New manual campaign'
    const header = campaign?.name ?? newCampaignHeader

    let stateBadge: JSX.Element

    if (!campaign) {
        stateBadge = <CampaignsIcon className="icon-inline campaign-actions-bar__campaign-icon text-muted mr-2" />
    } else if (campaign.closedAt) {
        stateBadge = (
            <span className="badge badge-danger mr-2">
                <CampaignsIcon className="icon-inline campaign-actions-bar__campaign-icon" /> Closed
            </span>
        )
    } else if (!campaign.publishedAt) {
        stateBadge = (
            <span className="badge badge-info mr-2">
                <CampaignsIcon className="icon-inline campaign-actions-bar__campaign-icon" /> Draft
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
        <div className="d-flex mb-2">
            <h2 className="m-0">
                {stateBadge}
                <span>
                    <Link to="/campaigns">Campaigns</Link>
                </span>
                <span className="text-muted d-inline-block mx-2">/</span>
                <span>{header}</span>
            </h2>
            <span className="flex-grow-1 d-flex justify-content-end align-items-center">
                {showSpinner && <LoadingSpinner className="mr-2" />}
                {campaign &&
                    showActionButtons &&
                    (editingCampaign ? (
                        <>
                            <button type="submit" className="btn btn-primary mr-1" disabled={mode === 'saving'}>
                                Save
                            </button>
                            <button type="reset" className="btn btn-secondary" disabled={mode === 'saving'}>
                                Cancel
                            </button>
                        </>
                    ) : (
                        <>
                            <button
                                type="button"
                                id="e2e-campaign-edit"
                                className="btn btn-secondary mr-1"
                                onClick={onEdit}
                                disabled={actionsDisabled}
                            >
                                Edit
                            </button>
                            {!campaign.closedAt && (
                                <CloseDeleteCampaignPrompt
                                    disabled={actionsDisabled}
                                    disabledTooltip="Cannot close while campaign is being created"
                                    message={
                                        <p>
                                            Close campaign <strong>{campaign.name}</strong>?
                                        </p>
                                    }
                                    changesetsCount={openChangesetsCount}
                                    buttonText="Close"
                                    onButtonClick={onClose}
                                    buttonClassName="btn-secondary mr-1"
                                />
                            )}
                            <CloseDeleteCampaignPrompt
                                disabled={actionsDisabled}
                                disabledTooltip="Cannot delete while campaign is being created"
                                message={
                                    <p>
                                        Delete campaign <strong>{campaign.name}</strong>?
                                    </p>
                                }
                                changesetsCount={openChangesetsCount}
                                buttonText="Delete"
                                onButtonClick={onDelete}
                                buttonClassName="btn-danger"
                            />
                        </>
                    ))}
            </span>
        </div>
    )
}
