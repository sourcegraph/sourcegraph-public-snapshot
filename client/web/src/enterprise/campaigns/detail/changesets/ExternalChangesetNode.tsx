import { ThemeProps } from '../../../../../../shared/src/theme'
import { Hoverifier } from '@sourcegraph/codeintellify'
import { RepoSpec, RevisionSpec, FileSpec, ResolvedRevisionSpec } from '../../../../../../shared/src/util/url'
import { HoverMerged } from '../../../../../../shared/src/api/client/types/hover'
import { ActionItemAction } from '../../../../../../shared/src/actions/ActionItem'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as H from 'history'
import React, { useState, useCallback } from 'react'
import { DiffStat } from '../../../../components/diff/DiffStat'
import { queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs } from '../backend'
import { ChangesetSpecType, ExternalChangesetFields } from '../../../../graphql-operations'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import { ChangesetStatusCell } from './ChangesetStatusCell'
import { ChangesetCheckStatusCell } from './ChangesetCheckStatusCell'
import { ChangesetReviewStatusCell } from './ChangesetReviewStatusCell'
import { ErrorAlert } from '../../../../components/alerts'
import { ChangesetFileDiff } from './ChangesetFileDiff'
import { ExternalChangesetInfoCell } from './ExternalChangesetInfoCell'
import { DownloadDiffButton } from './DownloadDiffButton'
import classNames from 'classnames'

export interface ExternalChangesetNodeProps extends ThemeProps {
    node: ExternalChangesetFields
    viewerCanAdminister: boolean
    history: H.History
    location: H.Location
    extensionInfo?: {
        hoverifier: Hoverifier<RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec, HoverMerged, ActionItemAction>
    } & ExtensionsControllerProps
    /** For testing only. */
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
    /** For testing only. */
    expandByDefault?: boolean
}

export const ExternalChangesetNode: React.FunctionComponent<ExternalChangesetNodeProps> = ({
    node,
    viewerCanAdminister,
    isLightTheme,
    history,
    location,
    extensionInfo,
    queryExternalChangesetWithFileDiffs,
    expandByDefault,
}) => {
    const [isExpanded, setIsExpanded] = useState(expandByDefault ?? false)
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
                className="btn btn-icon test-campaigns-expand-changeset d-none d-sm-block"
                aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                onClick={toggleIsExpanded}
            >
                {isExpanded ? (
                    <ChevronDownIcon className="icon-inline" aria-label="Close section" />
                ) : (
                    <ChevronRightIcon className="icon-inline" aria-label="Expand section" />
                )}
            </button>
            <ChangesetStatusCell
                changeset={node}
                className="p-2 align-self-stretch external-changeset-node__state d-block d-sm-flex"
            />
            <ExternalChangesetInfoCell
                node={node}
                viewerCanAdminister={viewerCanAdminister}
                className="p-2 align-self-stretch external-changeset-node__information"
            />
            <div
                className={classNames(
                    'd-flex d-md-none justify-content-start external-changeset-node__statuses',
                    (node.checkState || node.reviewState || node.diffStat) && 'p-2'
                )}
            >
                {node.checkState && <ChangesetCheckStatusCell checkState={node.checkState} className="mr-3" />}
                {node.reviewState && <ChangesetReviewStatusCell reviewState={node.reviewState} className="mr-3" />}
                {node.diffStat && <DiffStat {...node.diffStat} expandedCounts={true} separateLines={true} />}
            </div>
            <span className={classNames('align-self-stretch d-none d-md-flex', node.checkState && 'p-2')}>
                {node.checkState && <ChangesetCheckStatusCell checkState={node.checkState} />}
            </span>
            <span className={classNames('align-self-stretch d-none d-md-flex', node.reviewState && 'p-2')}>
                {node.reviewState && <ChangesetReviewStatusCell reviewState={node.reviewState} />}
            </span>
            <div
                className={classNames(
                    'align-self-stretch d-none d-md-flex justify-content-center',
                    node.diffStat && 'p-2'
                )}
            >
                {node.diffStat && <DiffStat {...node.diffStat} expandedCounts={true} separateLines={true} />}
            </div>
            {/* The button for expanding the information used on xs devices. */}
            <button
                type="button"
                aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                onClick={toggleIsExpanded}
                className="external-changeset-node__show-details btn btn-outline-secondary d-block d-sm-none test-campaigns-expand-changeset"
            >
                {isExpanded ? (
                    <ChevronDownIcon className="icon-inline" aria-label="Close section" />
                ) : (
                    <ChevronRightIcon className="icon-inline" aria-label="Expand section" />
                )}{' '}
                {isExpanded ? 'Hide' : 'Show'} details
            </button>
            {isExpanded && (
                <>
                    <div className="external-changeset-node__bg-expanded align-self-stretch" />
                    <div className="external-changeset-node__expanded-section external-changeset-node__bg-expanded p-2">
                        {node.currentSpec?.type === ChangesetSpecType.BRANCH && (
                            <DownloadDiffButton changesetID={node.id} />
                        )}
                        {node.error && <ErrorAlert error={node.error} history={history} />}
                        <ChangesetFileDiff
                            changesetID={node.id}
                            isLightTheme={isLightTheme}
                            history={history}
                            location={location}
                            repositoryID={node.repository.id}
                            repositoryName={node.repository.name}
                            extensionInfo={extensionInfo}
                            queryExternalChangesetWithFileDiffs={queryExternalChangesetWithFileDiffs}
                            updateOnChange={node.updatedAt}
                        />
                    </div>
                </>
            )}
        </>
    )
}
