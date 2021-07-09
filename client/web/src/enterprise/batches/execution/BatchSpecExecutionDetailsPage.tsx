import { isEqual } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import ErrorIcon from 'mdi-react/ErrorIcon'
import ProgressClockIcon from 'mdi-react/ProgressClockIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'
import React, { useMemo } from 'react'
import { delay, distinctUntilChanged, repeatWhen } from 'rxjs/operators'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { BatchSpecExecutionState } from '@sourcegraph/shared/src/graphql-operations'
import { isDefined } from '@sourcegraph/shared/src/util/types'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { BatchChangesIcon } from '../../../batches/icons'
import { ErrorAlert } from '../../../components/alerts'
import { HeroPage } from '../../../components/HeroPage'
import { PageTitle } from '../../../components/PageTitle'
import { Timeline } from '../../../components/Timeline'
import { BatchSpecExecutionFields, Scalars } from '../../../graphql-operations'
import { BatchSpec } from '../BatchSpec'

import { fetchBatchSpecExecution as _fetchBatchSpecExecution } from './backend'

export interface BatchSpecExecutionDetailsPageProps {
    executionID: Scalars['ID']

    /** For testing only. */
    fetchBatchSpecExecution?: typeof _fetchBatchSpecExecution
    /** For testing only. */
    now?: () => Date
}

export const BatchSpecExecutionDetailsPage: React.FunctionComponent<BatchSpecExecutionDetailsPageProps> = ({
    executionID,
    now = () => new Date(),
    fetchBatchSpecExecution = _fetchBatchSpecExecution,
}) => {
    const batchSpecExecution: BatchSpecExecutionFields | null | undefined = useObservable(
        useMemo(
            () =>
                fetchBatchSpecExecution(executionID).pipe(
                    repeatWhen(notifier => notifier.pipe(delay(2500))),
                    distinctUntilChanged((a, b) => isEqual(a, b))
                ),
            [fetchBatchSpecExecution, executionID]
        )
    )

    // Is loading.
    if (batchSpecExecution === undefined) {
        return (
            <div className="text-center">
                <LoadingSpinner className="icon-inline mx-auto my-4" />
            </div>
        )
    }
    // Is not found.
    if (batchSpecExecution === null) {
        return <HeroPage icon={AlertCircleIcon} title="Execution not found" />
    }

    return (
        <>
            <PageTitle title="Batch spec execution" />
            <PageHeader
                path={[
                    {
                        icon: BatchChangesIcon,
                        to: '/batch-changes',
                    },
                    {
                        to: `${batchSpecExecution.namespace.url}/batch-changes`,
                        text: batchSpecExecution.namespace.namespaceName,
                    },
                    { text: 'Execution' },
                ]}
                className="mb-3"
            />

            {batchSpecExecution.failure && <ErrorAlert error={batchSpecExecution.failure} />}

            <h2>Input spec</h2>
            <Container className="mb-3">
                <BatchSpec originalInput={batchSpecExecution.inputSpec} />
            </Container>

            <h2>Timeline</h2>
            <ExecutionTimeline execution={batchSpecExecution} now={now} className="mb-3" />

            {batchSpecExecution.batchSpec && (
                <>
                    <h2>Execution result</h2>
                    <div className="alert alert-info d-flex justify-space-between align-items-center">
                        <span className="flex-grow-1">Batch spec has been created.</span>
                        <Link to={batchSpecExecution.batchSpec.applyURL} className="btn btn-primary">
                            Preview changes
                        </Link>
                    </div>
                </>
            )}
        </>
    )
}

interface ExecutionTimelineProps {
    execution: BatchSpecExecutionFields
    className?: string

    /** For testing only. */
    now?: () => Date
}

const ExecutionTimeline: React.FunctionComponent<ExecutionTimelineProps> = ({ execution, className, now }) => {
    const stages = useMemo(
        () => [
            { icon: <TimerSandIcon />, text: 'Queued', date: execution.createdAt, className: 'bg-success' },
            {
                icon: <ProgressClockIcon />,
                text: 'Began processing',
                date: execution.startedAt,
                className: 'bg-success',
            },

            execution.state === BatchSpecExecutionState.COMPLETED
                ? { icon: <CheckIcon />, text: 'Finished', date: execution.finishedAt, className: 'bg-success' }
                : { icon: <ErrorIcon />, text: 'Failed', date: execution.finishedAt, className: 'bg-danger' },
        ],
        [execution]
    )
    return <Timeline stages={stages.filter(isDefined)} now={now} className={className} />
}
