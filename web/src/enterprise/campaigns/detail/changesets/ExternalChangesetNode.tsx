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
import { ExternalChangesetFields, GitCommitFields } from '../../../../graphql-operations'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import { ChangesetStatusCell } from './ChangesetStatusCell'
import { ChangesetCheckStatusCell } from './ChangesetCheckStatusCell'
import { ChangesetReviewStatusCell } from './ChangesetReviewStatusCell'
import { ErrorAlert } from '../../../../components/alerts'
import { ChangesetFileDiff } from './ChangesetFileDiff'
import { ExternalChangesetInfoCell } from './ExternalChangesetInfoCell'
import { GitCommitNode } from '../../../../repo/commits/GitCommitNode'

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
}

export const ExternalChangesetNode: React.FunctionComponent<ExternalChangesetNodeProps> = ({
    node,
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

    const [commits, setCommits] = useState<GitCommitFields[]>()

    return (
        <>
            <button
                type="button"
                className="btn btn-icon test-campaigns-expand-changeset"
                aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                onClick={toggleIsExpanded}
            >
                {isExpanded ? (
                    <ChevronDownIcon className="icon-inline" aria-label="Close section" />
                ) : (
                    <ChevronRightIcon className="icon-inline" aria-label="Expand section" />
                )}
            </button>
            <ChangesetStatusCell changeset={node} />
            <ExternalChangesetInfoCell node={node} viewerCanAdminister={viewerCanAdminister} />
            <span>{node.checkState && <ChangesetCheckStatusCell checkState={node.checkState} />}</span>
            <span>{node.reviewState && <ChangesetReviewStatusCell reviewState={node.reviewState} />}</span>
            <div className="external-changeset-node__diffstat">
                {node.diffStat && <DiffStat {...node.diffStat} expandedCounts={true} />}
            </div>
            {isExpanded && (
                <div className="external-changeset-node__expanded-section bg-gray">
                    {node.error && <ErrorAlert error={node.error} history={history} />}
                    <ul className="list-group px-5 mb-3">
                        {commits?.map(commit => (
                            <li className="list-group-item bg-white" key={commit.oid}>
                                <GitCommitNode node={commit} compact={true} />
                            </li>
                        ))}
                    </ul>
                    <ChangesetFileDiff
                        changesetID={node.id}
                        isLightTheme={isLightTheme}
                        history={history}
                        location={location}
                        repositoryID={node.repository.id}
                        repositoryName={node.repository.name}
                        extensionInfo={extensionInfo}
                        setCommits={setCommits}
                        queryExternalChangesetWithFileDiffs={queryExternalChangesetWithFileDiffs}
                    />
                </div>
            )}
        </>
    )
}
