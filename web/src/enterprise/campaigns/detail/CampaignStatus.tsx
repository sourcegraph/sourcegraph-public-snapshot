import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ErrorMessage } from '../../../components/alerts'
import SyncIcon from 'mdi-react/SyncIcon'
import { pluralize } from '../../../../../shared/src/util/strings'

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

    let statusIndicator: JSX.Element | undefined
    switch (state) {
        case 'errored':
            statusIndicator = (
                <>
                    <div className="alert alert-danger my-4">
                        <h3 className="alert-heading mb-0">Creating changesets failed</h3>
                        <ul className="mt-2">
                            {status.errors.map((error, i) => (
                                <li className="mb-2" key={i}>
                                    <code>
                                        <ErrorMessage error={error} />
                                    </code>
                                </li>
                            ))}
                        </ul>
                        {campaign.viewerCanAdminister && (
                            <button type="button" className="btn btn-primary mb-0" onClick={onRetry}>
                                Retry
                            </button>
                        )}
                    </div>
                </>
            )
            break
        case 'processing':
            statusIndicator = (
                <>
                    <div className="alert alert-info mt-4">
                        <p>
                            <SyncIcon className="icon-inline" /> Creating {status.pendingCount + status.completedCount}{' '}
                            {pluralize('changeset', status.pendingCount + status.completedCount)} on code hosts...
                        </p>
                        <div className="progress mt-2 mb-1">
                            {/* we need to set the width to control the progress bar, so: */}
                            {/* eslint-disable-next-line react/forbid-dom-props */}
                            <div className="progress-bar" style={{ width: progress + '%' }}>
                                &nbsp;
                            </div>
                        </div>
                    </div>
                </>
            )
            break
        case 'closed':
            statusIndicator = (
                <div className="alert alert-secondary mt-2">
                    Campaign is closed. No changes can be made to this campaign anymore.
                </div>
            )
            break
    }

    return (
        <>
            {statusIndicator && <div>{statusIndicator}</div>}
            {isDraft && state !== 'closed' && (
                <>
                    <div className="d-flex align-items-center alert alert-warning my-4">
                        {campaign.viewerCanAdminister && (
                            <button type="button" className="btn btn-primary mb-0" onClick={onPublish}>
                                Publish campaign
                            </button>
                        )}
                        <p className="mb-0 ml-2">
                            Campaign is a draft.{' '}
                            {campaign.changesets.totalCount === 0
                                ? 'No changesets have'
                                : 'Only a subset of changesets has'}{' '}
                            been created on code hosts yet.
                        </p>
                    </div>
                </>
            )}
        </>
    )
}
