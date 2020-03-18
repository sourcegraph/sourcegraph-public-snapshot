import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import WarningIcon from 'mdi-react/WarningIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import { ErrorAlert } from '../../../components/alerts'
import InformationIcon from 'mdi-react/InformationIcon'
import { parseISO, isBefore, addMinutes } from 'date-fns'
import { CampaignsIcon } from '../icons'
import SyncIcon from 'mdi-react/SyncIcon'

export interface CampaignStatusProps {
    campaign: Pick<GQL.ICampaign, 'closedAt' | 'viewerCanAdminister' | 'publishedAt'> & {
        changesets: Pick<GQL.ICampaign['changesets'], 'totalCount'>
        status: Pick<GQL.ICampaign['status'], 'completedCount' | 'pendingCount' | 'errors' | 'state'>
    }

    /** Called when the "Publish campaign" button is clicked. */
    onPublish: () => void
    /** Called when the "Retry failed jobs" button is clicked. */
    onRetry: () => void
}

type CampaignState = 'closed' | 'errored' | 'processing' | 'completed'

/**
 * The status of a campaign's jobs, plus its closed state and errors.
 */
export const CampaignStatus: React.FunctionComponent<CampaignStatusProps> = ({ campaign, onPublish, onRetry }) => {
    const { status } = campaign

    /* For completed campaigns that have been published, hide the creation complete status 1 day after the time of publication */
    const creationCompletedLongAgo =
        status.state === GQL.BackgroundProcessState.COMPLETED &&
        !!campaign.publishedAt &&
        isBefore(parseISO(campaign.publishedAt), addMinutes(new Date(), 1))

    const progress = (status.completedCount / (status.pendingCount + status.completedCount)) * 100

    const isDraft = !campaign.publishedAt
    let state: CampaignState
    if (campaign.closedAt) {
        state = 'closed'
    } else if (campaign.status.state === GQL.BackgroundProcessState.ERRORED) {
        state = 'errored'
    } else if (campaign.status.state === GQL.BackgroundProcessState.PROCESSING) {
        state = 'processing'
    } else {
        state = 'completed'
    }

    let statusIndicatorComponent: JSX.Element | undefined
    switch (state) {
        case 'completed':
            // no completion status for drafts
            if (isDraft) {
                break
            }
            statusIndicatorComponent = (
                <>
                    <CampaignsIcon className="icon-inline text-success mr-1" /> Campaign is open
                </>
            )
            break
        case 'errored':
            statusIndicatorComponent = (
                <>
                    <AlertCircleIcon className="icon-inline text-danger mr-1" />
                    Error creating campaign
                </>
            )
            break
        case 'processing':
            statusIndicatorComponent = (
                <>
                    <SyncIcon className="icon-inline text-info mr-1" />
                    Campaign processing
                </>
            )
            break
        case 'closed':
            statusIndicatorComponent = (
                <div className="alert alert-info mb-0">
                    <WarningIcon className="icon-inline mr-1" /> Campaign is a closed. No changes can be made to this
                    campaign anymore.
                </div>
            )
            break
    }

    return (
        <>
            {statusIndicatorComponent && <div>{statusIndicatorComponent}</div>}
            {isDraft && state !== 'closed' && (
                <>
                    <div className="d-flex alert alert-info mb-0 mt-2">
                        <InformationIcon className="icon-inline mr-1" /> Campaign is a draft.{' '}
                        {campaign.changesets.totalCount === 0
                            ? 'No changesets have'
                            : 'Only a subset of changesets has'}{' '}
                        been created on code hosts yet.
                    </div>
                    {campaign.viewerCanAdminister && (
                        <button type="button" className="btn btn-primary mt-2" onClick={onPublish}>
                            Publish campaign
                        </button>
                    )}
                </>
            )}
            {state === 'processing' && (
                <div>
                    <div className="progress mt-2 mb-1">
                        {/* we need to set the width to control the progress bar, so: */}
                        {/* eslint-disable-next-line react/forbid-dom-props */}
                        <div className="progress-bar" style={{ width: progress + '%' }}>
                            &nbsp;
                        </div>
                    </div>
                    <p>
                        Creating changes: {status.completedCount} / {status.pendingCount + status.completedCount}
                    </p>
                </div>
            )}
            {state === 'completed' && status.pendingCount + status.completedCount > 0 && !creationCompletedLongAgo && (
                <div className="d-flex mt-2">
                    <CheckCircleIcon className="icon-inline text-success mr-1" /> Creation completed
                </div>
            )}
            {status.errors.map((error, i) => (
                // There is no other suitable key, so:
                // eslint-disable-next-line react/no-array-index-key
                <ErrorAlert error={error} className="mt-2 mb-0" key={i} />
            ))}
            {state === 'errored' && campaign.viewerCanAdminister && (
                <button type="button" className="btn btn-primary mt-2" onClick={onRetry}>
                    Retry failed jobs
                </button>
            )}
        </>
    )
}
