import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import WarningIcon from 'mdi-react/WarningIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import { ErrorAlert } from '../../../components/alerts'
import { parseISO, addDays, isBefore } from 'date-fns'

interface Props {
    campaign:
        | Pick<GQL.ICampaign, '__typename' | 'closedAt' | 'publishedAt' | 'createdAt'>
        | Pick<GQL.ICampaignPlan, '__typename'>

    /** The campaign status. */
    status: Omit<GQL.IBackgroundProcessStatus, '__typename'>

    /** Called when the "Retry failed jobs" button is clicked. */
    onRetry: () => void
}

/**
 * The status of a campaign's jobs, plus its closed state and errors.
 */
export const CampaignStatus: React.FunctionComponent<Props> = ({ campaign, status, onRetry }) => {
    /* For completed campaigns that have been published, hide the creation complete status 1 day after the time of publication */
    const creationCompletedLongAgo =
        status.state === GQL.BackgroundProcessState.COMPLETED &&
        campaign.__typename === 'Campaign' &&
        campaign.publishedAt &&
        isBefore(parseISO(campaign.publishedAt), addDays(new Date(), 1))
    return (
        <>
            {status.state === GQL.BackgroundProcessState.PROCESSING && (
                <div className="d-flex mt-3 e2e-preview-loading">
                    <LoadingSpinner className="icon-inline" />{' '}
                    <span data-tooltip="Computing changesets">
                        {status.completedCount} / {status.pendingCount + status.completedCount}
                    </span>
                </div>
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
            {status.state === GQL.BackgroundProcessState.ERRORED && campaign?.__typename === 'Campaign' && (
                <button type="button" className="btn btn-primary mb-2" onClick={onRetry}>
                    Retry failed jobs
                </button>
            )}
        </>
    )
}
