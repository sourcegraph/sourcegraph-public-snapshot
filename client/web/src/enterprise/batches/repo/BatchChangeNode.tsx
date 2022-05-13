import React, { useState, useEffect } from 'react'

import classNames from 'classnames'
import * as H from 'history'

import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Link, Button, Typography } from '@sourcegraph/wildcard'

import { Timestamp } from '../../../components/time/Timestamp'
import { RepoBatchChange } from '../../../graphql-operations'
import { queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs } from '../detail/backend'
import { ChangesetNode } from '../detail/changesets/ChangesetNode'

import { MAX_CHANGESETS_COUNT } from './backend'

import styles from './BatchChangeNode.module.scss'

export interface BatchChangeNodeProps extends ThemeProps {
    node: RepoBatchChange
    viewerCanAdminister: boolean
    history: H.History
    location: H.Location
    /** For testing purposes. */
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
    /** For testing purposes. */
    expandByDefault?: boolean
    /** Used for testing purposes. Sets the current date */
    now?: () => Date
}

export const BatchChangeNode: React.FunctionComponent<React.PropsWithChildren<BatchChangeNodeProps>> = ({
    node: initialNode,
    now = () => new Date(),
    ...props
}) => {
    const [node, setNode] = useState(initialNode)
    useEffect(() => {
        setNode(initialNode)
    }, [initialNode])

    const moreChangesetsIndicator =
        node.changesets.totalCount > MAX_CHANGESETS_COUNT ? (
            <div className={classNames(styles.nodeFullWidth, 'text-center mt-2')}>
                <small>
                    <span>
                        {node.changesets.totalCount} changesets total (showing first {MAX_CHANGESETS_COUNT})
                    </span>
                </small>
                <Button
                    className="d-block"
                    to={node.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    variant="link"
                    size="sm"
                    as={Link}
                >
                    See all
                </Button>
            </div>
        ) : null

    return (
        <>
            <span className={styles.nodeSeparator} />
            <div className={styles.nodeFullWidth}>
                <div className="mt-1 mb-2 d-md-flex d-block align-items-baseline">
                    <Typography.H2 className="m-0 d-md-inline-block d-block">
                        <div className="d-md-inline-block d-block">
                            <Link
                                className="text-muted test-batches-namespace-link"
                                to={`${node.namespace.url}/batch-changes`}
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                {node.namespace.namespaceName}
                            </Link>
                            <span className="text-muted d-inline-block mx-1">/</span>
                        </div>
                        <Link
                            className="test-batches-link mr-2"
                            to={node.url}
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            {node.name}
                        </Link>
                    </Typography.H2>
                    <small className="text-muted d-sm-block">
                        created <Timestamp date={node.createdAt} now={now} />
                    </small>
                </div>
            </div>
            {node.changesets.nodes.map(changeset => (
                <ChangesetNode {...props} key={changeset.id} node={changeset} separator={null} />
            ))}
            {moreChangesetsIndicator}
            <div className={styles.nodeBottomSpacer} />
        </>
    )
}
