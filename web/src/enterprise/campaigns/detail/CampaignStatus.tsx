import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import WarningIcon from 'mdi-react/WarningIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import { ErrorAlert } from '../../../components/alerts'
import InformationIcon from 'mdi-react/InformationIcon'
import { parseISO, isBefore, addMinutes } from 'date-fns'

interface Props {
    campaign:
        | Pick<GQL.ICampaign, '__typename' | 'closedAt' | 'viewerCanAdminister' | 'publishedAt' | 'changesets'>
        | Pick<GQL.ICampaignPlan, '__typename'>

    /** The campaign status. */
    status: Omit<GQL.IBackgroundProcessStatus, '__typename'>

    /** Called when the "Publish campaign" button is clicked. */
    onPublish: () => void
    /** Called when the "Retry failed jobs" button is clicked. */
    onRetry: () => void
}

/**
 * The status of a campaign's jobs, plus its closed state and errors.
 */
export const CampaignStatus: React.FunctionComponent<Props> = ({ campaign, status, onPublish, onRetry }) => {
    /* For completed campaigns that have been published, hide the creation complete status 1 day after the time of publication */
    const creationCompletedLongAgo =
        status.state === GQL.BackgroundProcessState.COMPLETED &&
        campaign.__typename === 'Campaign' &&
        !!campaign.publishedAt &&
        isBefore(parseISO(campaign.publishedAt), addMinutes(new Date(), 1))
    const progress = (status.completedCount / (status.pendingCount + status.completedCount)) * 100
    return (
        <>
            {status.state === GQL.BackgroundProcessState.PROCESSING && (
                <div className="mt-3 e2e-preview-loading">
                    <div className="progress mb-1">
                        {/* we need to set the width to control the progress bar, so: */}
                        {/* eslint-disable-next-line react/forbid-dom-props */}
                        <div className="progress-bar" style={{ width: progress + '%' }}>
                            &nbsp;
                        </div>
                    </div>
                    <p>
                        {campaign.__typename === 'CampaignPlan' ? 'Computing' : 'Creating'} changes:{' '}
                        {status.completedCount} / {status.pendingCount + status.completedCount}
                    </p>
                </div>
            )}
            {campaign.__typename === 'Campaign' && !campaign.closedAt && !campaign.publishedAt && (
                <>
                    <div className="d-flex my-3">
                        <InformationIcon className="icon-inline text-info mr-1" /> Campaign is a draft.{' '}
                        {campaign.changesets.totalCount === 0
                            ? 'No changesets have'
                            : 'Only a subset of changesets has'}{' '}
                        been created on code hosts yet.
                    </div>
                    {campaign.viewerCanAdminister && (
                        <button type="button" className="mb-3 btn btn-primary" onClick={onPublish}>
                            Publish campaign
                        </button>
                    )}
                </>
            )}
            {campaign.__typename === 'Campaign' && campaign.closedAt ? (
                <div className="d-flex my-3">
                    <WarningIcon className="icon-inline text-warning mr-1" /> Campaign is closed
                </div>
            ) : (
                status.pendingCount + status.completedCount > 0 &&
                status.state !== GQL.BackgroundProcessState.PROCESSING &&
                !creationCompletedLongAgo && (
                    <div className="d-flex my-3">
                        {status.state === GQL.BackgroundProcessState.COMPLETED && (
                            <CheckCircleIcon className="icon-inline text-success mr-1 e2e-preview-success" />
                        )}
                        {status.state === GQL.BackgroundProcessState.ERRORED && (
                            <AlertCircleIcon className="icon-inline text-danger mr-1" />
                        )}{' '}
                        {campaign.__typename === 'Campaign' ? 'Creation' : 'Preview'} {status.state.toLocaleLowerCase()}
                    </div>
                )
            )}
            {status.errors.map((error, i) => (
                // There is no other suitable key, so:
                // eslint-disable-next-line react/no-array-index-key
                <ErrorAlert error={error} className="mt-3" key={i} />
            ))}
            {status.state === GQL.BackgroundProcessState.ERRORED &&
                campaign?.__typename === 'Campaign' &&
                !campaign.closedAt &&
                campaign.viewerCanAdminister && (
                    <button type="button" className="btn btn-primary mb-2" onClick={onRetry}>
                        Retry failed jobs
                    </button>
                )}
        </>
    )
}
