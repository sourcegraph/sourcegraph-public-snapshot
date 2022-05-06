import React, { useState, useCallback } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { HoverMerged } from '@sourcegraph/client-api'
import { Hoverifier } from '@sourcegraph/codeintellify'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { RepoSpec, RevisionSpec, FileSpec, ResolvedRevisionSpec } from '@sourcegraph/shared/src/util/url'
import { Button, Icon } from '@sourcegraph/wildcard'

import { DiffStatStack } from '../../../components/diff/DiffStat'
import { ExternalChangesetFields } from '../../../graphql-operations'
import { queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs } from '../detail/backend'
import { ChangesetCheckStatusCell } from '../detail/changesets/ChangesetCheckStatusCell'
import { ChangesetFileDiff } from '../detail/changesets/ChangesetFileDiff'
import { ChangesetReviewStatusCell } from '../detail/changesets/ChangesetReviewStatusCell'
import { ExternalChangesetInfoCell } from '../detail/changesets/ExternalChangesetInfoCell'

import { ChangesetCloseActionClose, ChangesetCloseActionKept } from './ChangesetCloseAction'

import styles from './ExternalChangesetCloseNode.module.scss'

export interface ExternalChangesetCloseNodeProps extends ThemeProps {
    node: ExternalChangesetFields
    willClose: boolean
    viewerCanAdminister: boolean
    history: H.History
    location: H.Location
    extensionInfo?: {
        hoverifier: Hoverifier<RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec, HoverMerged, ActionItemAction>
    } & ExtensionsControllerProps
    /** For testing only. */
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
}

export const ExternalChangesetCloseNode: React.FunctionComponent<
    React.PropsWithChildren<ExternalChangesetCloseNodeProps>
> = ({
    node,
    willClose,
    viewerCanAdminister,
    isLightTheme,
    history,
    location,
    extensionInfo,
    queryExternalChangesetWithFileDiffs,
}) => {
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
                {isExpanded ? (
                    <Icon aria-label="Close section" as={ChevronDownIcon} />
                ) : (
                    <Icon aria-label="Expand section" as={ChevronRightIcon} />
                )}
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
                {isExpanded ? (
                    <Icon aria-label="Close section" as={ChevronDownIcon} />
                ) : (
                    <Icon aria-label="Expand section" as={ChevronRightIcon} />
                )}{' '}
                {isExpanded ? 'Hide' : 'Show'} details
            </Button>
            {isExpanded && (
                <div className={classNames(styles.externalChangesetCloseNodeExpandedSection, 'p-2')}>
                    {node.error && <ErrorAlert error={node.error} />}
                    <ChangesetFileDiff
                        changesetID={node.id}
                        isLightTheme={isLightTheme}
                        history={history}
                        location={location}
                        repositoryID={node.repository.id}
                        repositoryName={node.repository.name}
                        extensionInfo={extensionInfo}
                        queryExternalChangesetWithFileDiffs={queryExternalChangesetWithFileDiffs}
                    />
                </div>
            )}
        </>
    )
}
