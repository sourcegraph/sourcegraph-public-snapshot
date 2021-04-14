import * as H from 'history'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React, { useState, useCallback } from 'react'

import { Hoverifier } from '@sourcegraph/codeintellify'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { HoverMerged } from '@sourcegraph/shared/src/api/client/types/hover'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { RepoSpec, RevisionSpec, FileSpec, ResolvedRevisionSpec } from '@sourcegraph/shared/src/util/url'

import { ErrorAlert } from '../../../components/alerts'
import { DiffStat } from '../../../components/diff/DiffStat'
import { ExternalChangesetFields } from '../../../graphql-operations'
import { queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs } from '../detail/backend'
import { ChangesetCheckStatusCell } from '../detail/changesets/ChangesetCheckStatusCell'
import { ChangesetFileDiff } from '../detail/changesets/ChangesetFileDiff'
import { ChangesetReviewStatusCell } from '../detail/changesets/ChangesetReviewStatusCell'
import { ExternalChangesetInfoCell } from '../detail/changesets/ExternalChangesetInfoCell'

import { ChangesetCloseActionClose, ChangesetCloseActionKept } from './ChangesetCloseAction'

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

export const ExternalChangesetCloseNode: React.FunctionComponent<ExternalChangesetCloseNodeProps> = ({
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
            <button
                type="button"
                className="btn btn-icon test-batches-expand-changeset d-none d-sm-block"
                aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                onClick={toggleIsExpanded}
            >
                {isExpanded ? (
                    <ChevronDownIcon className="icon-inline" aria-label="Close section" />
                ) : (
                    <ChevronRightIcon className="icon-inline" aria-label="Expand section" />
                )}
            </button>
            {willClose ? (
                <ChangesetCloseActionClose className="external-changeset-close-node__action" />
            ) : (
                <ChangesetCloseActionKept className="external-changeset-close-node__action" />
            )}
            <ExternalChangesetInfoCell
                node={node}
                viewerCanAdminister={viewerCanAdminister}
                className="external-changeset-close-node__information"
            />
            <div className="d-flex d-md-none justify-content-start external-changeset-close-node__statuses">
                {node.checkState && <ChangesetCheckStatusCell checkState={node.checkState} className="mr-3" />}
                {node.reviewState && <ChangesetReviewStatusCell reviewState={node.reviewState} className="mr-3" />}
                {node.diffStat && <DiffStat {...node.diffStat} expandedCounts={true} separateLines={true} />}
            </div>
            <span className="d-none d-md-inline">
                {node.checkState && <ChangesetCheckStatusCell checkState={node.checkState} />}
            </span>
            <span className="d-none d-md-inline">
                {node.reviewState && <ChangesetReviewStatusCell reviewState={node.reviewState} />}
            </span>
            <div className="d-none d-md-flex justify-content-center">
                {node.diffStat && <DiffStat {...node.diffStat} expandedCounts={true} separateLines={true} />}
            </div>
            {/* The button for expanding the information used on xs devices. */}
            <button
                type="button"
                aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                onClick={toggleIsExpanded}
                className="external-changeset-close-node__show-details btn btn-outline-secondary d-block d-sm-none test-batches-expand-changeset"
            >
                {isExpanded ? (
                    <ChevronDownIcon className="icon-inline" aria-label="Close section" />
                ) : (
                    <ChevronRightIcon className="icon-inline" aria-label="Expand section" />
                )}{' '}
                {isExpanded ? 'Hide' : 'Show'} details
            </button>
            {isExpanded && (
                <div className="external-changeset-close-node__expanded-section p-2">
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
