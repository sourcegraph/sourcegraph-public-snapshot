import * as H from 'history'
import React, { useState, useEffect } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { Timestamp } from '../../../components/time/Timestamp'
import { RepoBatchChange } from '../../../graphql-operations'
import { queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs } from '../detail/backend'
import { ChangesetNode } from '../detail/changesets/ChangesetNode'

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

export const BatchChangeNode: React.FunctionComponent<BatchChangeNodeProps> = ({
    node: initialNode,
    now = () => new Date(),
    ...props
}) => {
    const [node, setNode] = useState(initialNode)
    useEffect(() => {
        setNode(initialNode)
    }, [initialNode])

    return (
        <>
            <span className={styles.nodeSeparator} />
            <div className={styles.nodeFullWidth}>
                <div className="mt-1 mb-2 d-md-flex d-block align-items-baseline">
                    <h2 className="m-0 d-md-inline-block d-block">
                        <div className="d-md-inline-block d-block">
                            <Link
                                className="text-muted test-batches-namespace-link"
                                to={`${node.namespace.url}/batch-changes`}
                            >
                                {node.namespace.namespaceName}
                            </Link>
                            <span className="text-muted d-inline-block mx-1">/</span>
                        </div>
                        <Link className="test-batches-link mr-2" to={node.url}>
                            {node.name}
                        </Link>
                    </h2>
                    <small className="text-muted d-sm-block">
                        created <Timestamp date={node.createdAt} now={now} />
                    </small>
                </div>
            </div>
            {node.changesets.nodes.map(changeset => (
                <ChangesetNode {...props} key={changeset.id} node={changeset} separator={null} />
            ))}
            <div className={styles.nodeBottomSpacer} />
        </>
    )
}
