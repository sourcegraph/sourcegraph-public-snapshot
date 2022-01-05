import classNames from 'classnames'
import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'
import { Badge } from '@sourcegraph/wildcard'

import { Timestamp } from '../../../components/time/Timestamp'
import { BatchChangeState, ListBatchChange } from '../../../graphql-operations'
import {
    ChangesetStatusOpen,
    ChangesetStatusClosed,
    ChangesetStatusMerged,
} from '../detail/changesets/ChangesetStatusCell'

import styles from './BatchChangeNode.module.scss'

export interface BatchChangeNodeProps {
    node: ListBatchChange
    /** Used for testing purposes. Sets the current date */
    now?: () => Date
    displayNamespace: boolean
}

const StateBadge: React.FunctionComponent<{ state: BatchChangeState }> = ({ state }) => {
    switch (state) {
        case BatchChangeState.OPEN:
            return (
                <Badge variant="success" className={classNames(styles.batchChangeNodeBadge, 'text-uppercase')}>
                    Open
                </Badge>
            )
        case BatchChangeState.CLOSED:
            return (
                <Badge variant="danger" className={classNames(styles.batchChangeNodeBadge, 'text-uppercase')}>
                    Closed
                </Badge>
            )
        case BatchChangeState.DRAFT:
        default:
            return (
                <Badge variant="secondary" className={classNames(styles.batchChangeNodeBadge, 'text-uppercase')}>
                    Draft
                </Badge>
            )
    }
}

/**
 * An item in the list of batch changes.
 */
export const BatchChangeNode: React.FunctionComponent<BatchChangeNodeProps> = ({
    node,
    now = () => new Date(),
    displayNamespace,
}) => (
    <>
        <span className={styles.batchChangeNodeSeparator} />
        <StateBadge state={node.state} />
        <div className={styles.batchChangeNodeContent}>
            <div className="m-0 d-md-flex d-block align-items-baseline">
                <h3 className={classNames(styles.batchChangeNodeTitle, 'm-0 d-md-inline-block d-block')}>
                    {displayNamespace && (
                        <div className="d-md-inline-block d-block">
                            <Link
                                className="text-muted test-batches-namespace-link"
                                to={`${node.namespace.url}/batch-changes`}
                            >
                                {node.namespace.namespaceName}
                            </Link>
                            <span className="text-muted d-inline-block mx-1">/</span>
                        </div>
                    )}
                    <Link
                        className="test-batches-link mr-2"
                        to={`${node.url}${node.state === BatchChangeState.DRAFT ? '/edit' : ''}`}
                    >
                        {node.name}
                    </Link>
                </h3>
                <small className="text-muted d-sm-block">
                    created <Timestamp date={node.createdAt} now={now} />
                </small>
            </div>
            <Markdown
                className={classNames('text-truncate text-muted d-none d-md-block', !node.description && 'font-italic')}
                dangerousInnerHTML={
                    node.description ? renderMarkdown(node.description, { plainText: true }) : 'No description'
                }
            />
        </div>
        {node.state !== BatchChangeState.DRAFT && (
            <>
                <ChangesetStatusOpen
                    className="d-block d-sm-flex"
                    label={<span className="text-muted">{node.changesetsStats.open} open</span>}
                />
                <ChangesetStatusClosed
                    className="d-block d-sm-flex text-center"
                    label={<span className="text-muted">{node.changesetsStats.closed} closed</span>}
                />
                <ChangesetStatusMerged
                    className="d-block d-sm-flex"
                    label={<span className="text-muted">{node.changesetsStats.merged} merged</span>}
                />
            </>
        )}
    </>
)
