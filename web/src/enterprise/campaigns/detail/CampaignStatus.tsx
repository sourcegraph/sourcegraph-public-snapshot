import * as React from 'react'
import { ICampaign, ICampaignPlan } from '../../../../../shared/src/graphql/schema'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import { ErrorAlert } from '../../../components/alerts'
import WarningIcon from 'mdi-react/WarningIcon'

interface Props {
    campaign: ICampaign | ICampaignPlan
    onRetry: React.MouseEventHandler
}

export const CampaignStatus: React.FunctionComponent<Props> = ({ campaign, onRetry }) => {
    const status = campaign.__typename === 'CampaignPlan' ? campaign.status : campaign.changesetCreationStatus
    return (
        <>
            {status.state === 'PROCESSING' && (
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
                status.state !== 'PROCESSING' && (
                    <div className="d-flex my-3">
                        {status.state === 'COMPLETED' && (
                            <CheckCircleIcon className="icon-inline text-success mr-1 e2e-preview-success" />
                        )}
                        {status.state === 'ERRORED' && <AlertCircleIcon className="icon-inline text-danger mr-1" />}{' '}
                        {/* Status asserts on campaign being set, this will never be null */}
                        {campaign.__typename === 'Campaign' ? 'Creation' : 'Preview'} {status.state.toLocaleLowerCase()}
                    </div>
                )
            )}
            {status.errors.map((error, i) => (
                <ErrorAlert error={error} className="mt-3" key={i} />
            ))}
            {status.state === 'ERRORED' && campaign.__typename === 'Campaign' && !campaign.closedAt && (
                <button type="button" className="btn btn-primary mb-2" onClick={onRetry}>
                    Retry failed jobs
                </button>
            )}
        </>
    )
}
