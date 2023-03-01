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
} from '@sourcegraph/wildcard'

import { RepoEmbeddingJobFields, RepoEmbeddingJobState } from '../../../graphql-operations'

import styles from './RepoEmbeddingJobNode.module.scss'

interface RepoEmbeddingJobNodeProps extends RepoEmbeddingJobFields {}

export const RepoEmbeddingJobNode: FC<RepoEmbeddingJobNodeProps> = ({
    state,
    repo,
    revision,
    finishedAt,
    queuedAt,
    startedAt,
    failureMessage,
}) => (
    <li className="list-group-item p-2">
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
                        finishedAt={finishedAt}
                        queuedAt={queuedAt}
                        startedAt={startedAt}
                        failureMessage={failureMessage}
                    />
                </div>
            </div>
        </div>
    </li>
)

const RepoEmbeddingJobExecutionInfo: FC<
    Pick<RepoEmbeddingJobFields, 'state' | 'finishedAt' | 'failureMessage' | 'queuedAt' | 'startedAt'>
> = ({ state, finishedAt, queuedAt, startedAt, failureMessage }) => {
    const [isPopoverOpen, setIsPopoverOpen] = useState(false)
    return (
        <>
            {state === RepoEmbeddingJobState.COMPLETED && finishedAt && (
                <small>
                    Embedding completed successfully <Timestamp date={finishedAt} />.
                </small>
            )}
            {state === RepoEmbeddingJobState.CANCELED && <small>Embedding was canceled.</small>}
            {state === RepoEmbeddingJobState.QUEUED && (
                <small>
                    Embedding was queued <Timestamp date={queuedAt} />.
                </small>
            )}
            {state === RepoEmbeddingJobState.PROCESSING && startedAt && (
                <small>
                    Embedding started processing <Timestamp date={startedAt} />.
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
