import React, { useState, useCallback } from 'react'

import { mdiChevronDown, mdiChevronUp } from '@mdi/js'
import classNames from 'classnames'

import { Button, Icon, ErrorAlert } from '@sourcegraph/wildcard'

import { DiffStatStack } from '../../../components/diff/DiffStat'
import type { ExternalChangesetFields } from '../../../graphql-operations'
import type { queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs } from '../detail/backend'
import { ChangesetCheckStatusCell } from '../detail/changesets/ChangesetCheckStatusCell'
import { ChangesetFileDiff } from '../detail/changesets/ChangesetFileDiff'
import { ChangesetReviewStatusCell } from '../detail/changesets/ChangesetReviewStatusCell'
import { ExternalChangesetInfoCell } from '../detail/changesets/ExternalChangesetInfoCell'

import { ChangesetCloseActionClose, ChangesetCloseActionKept } from './ChangesetCloseAction'

import styles from './ExternalChangesetCloseNode.module.scss'

export interface ExternalChangesetCloseNodeProps {
    node: ExternalChangesetFields
    willClose: boolean
    viewerCanAdminister: boolean
    /** For testing only. */
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
}

export const ExternalChangesetCloseNode: React.FunctionComponent<
    React.PropsWithChildren<ExternalChangesetCloseNodeProps>
> = ({ node, willClose, viewerCanAdminister, queryExternalChangesetWithFileDiffs }) => {
    const [isExpanded, setIsExpanded] = useState(false)
    const toggleIsExpanded = useCallback<React.MouseEventHandler<HTMLButtonElement>>(
        event => {
            event.preventDefault()
            setIsExpanded(!isExpanded)
        },
        [isExpanded]
    )

    return (
        <>
            <Button
                variant="icon"
                className="test-batches-expand-changeset d-none d-sm-block"
                aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                onClick={toggleIsExpanded}
            >
                <Icon aria-hidden={true} svgPath={isExpanded ? mdiChevronUp : mdiChevronDown} />
            </Button>
            {willClose ? (
                <ChangesetCloseActionClose className={styles.externalChangesetCloseNodeAction} />
            ) : (
                <ChangesetCloseActionKept className={styles.externalChangesetCloseNodeAction} />
            )}
            <ExternalChangesetInfoCell
                node={node}
                viewerCanAdminister={viewerCanAdminister}
                className={styles.externalChangesetCloseNodeInformation}
            />
            <div
                className={classNames(
                    styles.externalChangesetCloseNodeStatuses,
                    'd-flex d-md-none justify-content-start'
                )}
            >
                {node.checkState && <ChangesetCheckStatusCell checkState={node.checkState} className="mr-3" />}
                {node.reviewState && <ChangesetReviewStatusCell reviewState={node.reviewState} className="mr-3" />}
                {node.diffStat && <DiffStatStack {...node.diffStat} />}
            </div>
            <span className="d-none d-md-inline">
                {node.checkState && <ChangesetCheckStatusCell checkState={node.checkState} />}
            </span>
            <span className="d-none d-md-inline">
                {node.reviewState && <ChangesetReviewStatusCell reviewState={node.reviewState} />}
            </span>
            <div className="d-none d-md-flex justify-content-center">
                {node.diffStat && <DiffStatStack {...node.diffStat} />}
            </div>
            {/* The button for expanding the information used on xs devices. */}
            <Button
                aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                onClick={toggleIsExpanded}
                className={classNames(
                    styles.externalChangesetCloseNodeShowDetails,
                    'd-block d-sm-none test-batches-expand-changeset'
                )}
                outline={true}
                variant="secondary"
            >
                <Icon aria-hidden={true} svgPath={isExpanded ? mdiChevronUp : mdiChevronDown} />{' '}
                {isExpanded ? 'Hide' : 'Show'} details
            </Button>
            {isExpanded && (
                <div className={classNames(styles.externalChangesetCloseNodeExpandedSection, 'p-2')}>
                    {node.error && <ErrorAlert error={node.error} />}
                    <ChangesetFileDiff
                        changesetID={node.id}
                        repositoryID={node.repository.id}
                        repositoryName={node.repository.name}
                        queryExternalChangesetWithFileDiffs={queryExternalChangesetWithFileDiffs}
                    />
                </div>
            )}
        </>
    )
}
