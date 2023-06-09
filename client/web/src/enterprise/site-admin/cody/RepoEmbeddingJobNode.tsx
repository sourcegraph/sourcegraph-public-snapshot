import { FC, useState } from 'react'

import classNames from 'classnames'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import {
    Badge,
    Button,
    PopoverTrigger,
    PopoverContent,
    PopoverTail,
    Popover,
    Position,
    BadgeVariantType,
    Link,
    H4,
    Alert,
    Tooltip,
} from '@sourcegraph/wildcard'

import { RepoEmbeddingJobFields, RepoEmbeddingJobState } from '../../../graphql-operations'

import styles from './RepoEmbeddingJobNode.module.scss'

interface RepoEmbeddingJobNodeProps extends RepoEmbeddingJobFields {
    onCancel: (id: string) => void
}

export const RepoEmbeddingJobNode: FC<RepoEmbeddingJobNodeProps> = ({
    id,
    state,
    cancel,
    repo,
    revision,
    finishedAt,
    queuedAt,
    startedAt,
    failureMessage,
    filesScheduled,
    filesEmbedded,
    filesSkipped,
    onCancel,
}) => (
    <li className="list-group-item p-2">
        <div className="d-flex justify-content-between">
            <div className="d-flex align-items-center">
                <div className={styles.badgeWrapper}>
                    <RepoEmbeddingJobStateBadge state={state} />
                </div>
                <div className="d-flex flex-column ml-3">
                    {repo && revision ? (
                        <Link to={`${repo.url}@${revision.oid}`}>
                            {repo.name}@{revision.abbreviatedOID}
                        </Link>
                    ) : (
                        <div>Unknown repository</div>
                    )}
                    <div className="mt-1">
                        <RepoEmbeddingJobExecutionInfo
                            state={state}
                            cancel={cancel}
                            finishedAt={finishedAt}
                            queuedAt={queuedAt}
                            startedAt={startedAt}
                            failureMessage={failureMessage}
                            filesScheduled={filesScheduled}
                            filesEmbedded={filesEmbedded}
                            filesSkipped={filesSkipped}
                        />
                    </div>
                </div>
            </div>
            <div className="d-flex align-items-center">
                {state === RepoEmbeddingJobState.QUEUED || state === RepoEmbeddingJobState.PROCESSING ? (
                    <Tooltip content="Cancel repository embedding job">
                        <Button
                            aria-label="Cancel"
                            onClick={() => onCancel(id)}
                            variant="secondary"
                            size="sm"
                            disabled={cancel}
                        >
                            Cancel
                        </Button>
                    </Tooltip>
                ) : null}
            </div>
        </div>
    </li>
)

const RepoEmbeddingJobExecutionInfo: FC<
    Pick<
        RepoEmbeddingJobFields,
        | 'state'
        | 'cancel'
        | 'finishedAt'
        | 'failureMessage'
        | 'queuedAt'
        | 'startedAt'
        | 'filesScheduled'
        | 'filesEmbedded'
        | 'filesSkipped'
    >
> = ({
    state,
    cancel,
    finishedAt,
    queuedAt,
    startedAt,
    failureMessage,
    filesScheduled,
    filesEmbedded,
    filesSkipped,
}) => {
        const [isPopoverOpen, setIsPopoverOpen] = useState(false)
        const estimatedFinish = calculateEstimatedFinish(startedAt, filesScheduled, filesEmbedded, filesSkipped)

        return (
            <>
                {state === RepoEmbeddingJobState.COMPLETED && finishedAt && (
                    <small>
                        Completed embedding {filesEmbedded} files ({filesSkipped} skipped) <Timestamp date={finishedAt} />
                    </small>
                )}
                {state === RepoEmbeddingJobState.CANCELED && finishedAt && (
                    <small>
                        Stopped <Timestamp date={finishedAt} />
                    </small>
                )}
                {state === RepoEmbeddingJobState.QUEUED && (
                    <small>
                        {cancel ? (
                            'Cancelling ...'
                        ) : (
                            <>
                                Queued <Timestamp date={queuedAt} />
                            </>
                        )}
                    </small>
                )}
                {state === RepoEmbeddingJobState.PROCESSING && startedAt && (
                    <small>
                        {cancel ? (
                            'Cancelling ...'
                        ) : estimatedFinish ? (
                            <>
                                Expected to finish <Timestamp date={estimatedFinish} /> ({filesSkipped + filesEmbedded}/
                                {filesScheduled} files)
                            </>
                        ) : (
                            <>
                                Started processing <Timestamp date={startedAt} />
                            </>
                        )}
                    </small>
                )}
                {(state === RepoEmbeddingJobState.ERRORED || state === RepoEmbeddingJobState.FAILED) && failureMessage && (
                    <Popover isOpen={isPopoverOpen} onOpenChange={event => setIsPopoverOpen(event.isOpen)}>
                        <PopoverTrigger as={Button} variant="secondary" size="sm" aria-label="See errors">
                            See errors
                        </PopoverTrigger>

                        <PopoverContent position={Position.right} className={styles.errorContent}>
                            <Alert variant="danger" className={classNames('m-2', styles.alertOverflow)}>
                                <H4>Error embedding repository:</H4>
                                <div>{failureMessage}</div>
                            </Alert>
                        </PopoverContent>
                        <PopoverTail size="sm" />
                    </Popover>
                )}
            </>
        )
    }

function calculateEstimatedFinish(
    startedAt: string | null,
    filesScheduled: number,
    filesEmbedded: number,
    filesSkipped: number,
    now?: number
): Date | null {
    const currentTime = now ?? Date.now()
    if (!startedAt) {
        return null
    }
    const startTime = Date.parse(startedAt)
    if (filesScheduled === 0) {
        return null
    }
    const proportionFinished = (filesEmbedded + filesSkipped) / filesScheduled
    const timeElapsed = currentTime - startTime
    const estimatedTotalTime = timeElapsed / proportionFinished
    return new Date(startTime + estimatedTotalTime)
}

function getRepoEmbeddingJobStateBadgeVariant(state: RepoEmbeddingJobState): BadgeVariantType {
    switch (state) {
        case RepoEmbeddingJobState.COMPLETED:
            return 'success'
        case RepoEmbeddingJobState.CANCELED:
            return 'warning'
        case RepoEmbeddingJobState.ERRORED:
        case RepoEmbeddingJobState.FAILED:
            return 'danger'
        case RepoEmbeddingJobState.QUEUED:
        case RepoEmbeddingJobState.PROCESSING:
            return 'info'
    }
}

const RepoEmbeddingJobStateBadge: React.FunctionComponent<{ state: RepoEmbeddingJobState }> = ({ state }) => (
    <Badge
        variant={getRepoEmbeddingJobStateBadgeVariant(state)}
        className={classNames('py-0 px-1 text-uppercase font-weight-normal')}
    >
        {state}
    </Badge>
)
