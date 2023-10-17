import { type ChangeEvent, type FC, type PropsWithChildren, useId } from 'react'

import { mdiAlertCircle, mdiCheckCircle, mdiHelp, mdiMoonNew, mdiTimerSand } from '@mdi/js'
import classNames from 'classnames'
import { timeFormat } from 'd3-time-format'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/branded'
import {
    Badge,
    Button,
    ErrorAlert,
    H3,
    Icon,
    Label,
    LoadingSpinner,
    Popover,
    PopoverContent,
    PopoverTail,
    PopoverTrigger,
} from '@sourcegraph/wildcard'

import { type InsightJob, InsightQueueItemState } from '../../../../../graphql-operations'
import { formatFilter } from '../job-filters'

import styles from './CodeInsightsJobCard.module.scss'

const formatDate = timeFormat('%Y-%m-%d %H:%M:%S')

interface CodeInsightsJobCardProps {
    job: InsightJob
    selected: boolean
    onSelectChange: (event: ChangeEvent<HTMLInputElement>) => void
}

export const CodeInsightsJobCard: FC<CodeInsightsJobCardProps> = props => {
    const {
        selected,
        job: {
            insightViewTitle,
            seriesLabel,
            seriesSearchQuery,
            backfillQueueStatus: {
                state,
                cost,
                errors,
                percentComplete,
                queuePosition,
                createdAt,
                startedAt,
                completedAt,
            },
        },
        onSelectChange,
    } = props

    const checkboxId = useId()

    const details = [
        queuePosition !== null && `Queue position: ${queuePosition}`,
        cost !== null && `Cost: ${cost}`,
        createdAt !== null && `Created at: ${formatDate(new Date(createdAt))}`,
        startedAt !== null && `Started at: ${formatDate(new Date(startedAt))}`,
        completedAt !== null && `Completed at: ${formatDate(new Date(completedAt))}`,
    ].filter(item => item)

    return (
        <li className={classNames(styles.insightJob, { [styles.insightJobActive]: selected })}>
            {/* eslint-disable-next-line react/forbid-elements */}
            <input
                id={checkboxId}
                type="checkbox"
                checked={selected}
                className={styles.insightJobCheckbox}
                onChange={onSelectChange}
            />
            <div className={styles.insightJobContent}>
                <header className={styles.insightJobHeader}>
                    {seriesLabel && (
                        <H3 as={Label} htmlFor={checkboxId} className={styles.insightJobTitle}>
                            {seriesLabel}
                        </H3>
                    )}

                    <small className="text-muted">From</small>
                    <Pill className={styles.insightJobSubtitle}>{insightViewTitle} insight</Pill>
                </header>

                <span className={styles.insightJobMainInfo}>
                    {percentComplete !== null && <span>{percentComplete}% completed</span>}
                    <span className={styles.insightJobQueryBlock}>
                        Series query:{' '}
                        <SyntaxHighlightedSearchQuery query={seriesSearchQuery} className={styles.insightJobQuery} />
                    </span>
                </span>

                {details.length > 0 && <small className="mt-1 text-muted">{details.join(', ')}</small>}
            </div>
            <div className={styles.insightJobState}>
                <span className={styles.insightJobStateIcon}>
                    <InsightJobStatusIcon status={state} className={StatusClasses[state]} />
                    {formatFilter(state)}
                </span>
                {errors && errors.length > 0 && <InsightJobErrors errors={errors} />}
            </div>
        </li>
    )
}

const Pill: FC<PropsWithChildren<{ className?: string }>> = props => (
    <Badge {...props} as="small" variant="secondary" className={classNames(styles.pill, props.className)} />
)

const StatusIcon: Record<InsightQueueItemState, string> = {
    [InsightQueueItemState.COMPLETED]: mdiCheckCircle,
    [InsightQueueItemState.FAILED]: mdiAlertCircle,
    [InsightQueueItemState.NEW]: mdiMoonNew,
    [InsightQueueItemState.QUEUED]: mdiTimerSand,
    [InsightQueueItemState.UNKNOWN]: mdiHelp,
    [InsightQueueItemState.PROCESSING]: '',
}

const StatusClasses: Record<InsightQueueItemState, string> = {
    [InsightQueueItemState.COMPLETED]: styles.insightJobStateCompleted,
    [InsightQueueItemState.FAILED]: styles.insightJobStateErrored,
    [InsightQueueItemState.NEW]: styles.insightJobStateQueued,
    [InsightQueueItemState.QUEUED]: styles.insightJobStateQueued,
    [InsightQueueItemState.PROCESSING]: '',
    [InsightQueueItemState.UNKNOWN]: '',
}

interface InsightJobStatusProps {
    status: InsightQueueItemState
    className?: string
}

const InsightJobStatusIcon: FC<InsightJobStatusProps> = props => {
    const { status, className } = props

    if (status === InsightQueueItemState.PROCESSING) {
        return <LoadingSpinner inline={false} className={className} />
    }

    return (
        <Icon
            svgPath={StatusIcon[status]}
            width={20}
            height={20}
            inline={false}
            className={className}
            aria-hidden={true}
        />
    )
}

interface InsightJobErrorsProps {
    errors: string[]
}

const InsightJobErrors: FC<InsightJobErrorsProps> = props => {
    const { errors } = props

    return (
        <Popover>
            <PopoverTrigger as={Button} size="sm" outline={true} variant="danger" className={styles.errorsTrigger}>
                Show errors log
            </PopoverTrigger>
            <PopoverContent className={styles.errorsContent} focusLocked={false}>
                {errors.map(error => (
                    <ErrorAlert key={error} error={error} className={styles.error} />
                ))}
            </PopoverContent>
            <PopoverTail size="sm" />
        </Popover>
    )
}
