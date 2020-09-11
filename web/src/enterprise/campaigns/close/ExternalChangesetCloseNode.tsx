import { Observer } from 'rxjs'
import { Hoverifier } from '@sourcegraph/codeintellify'
import * as H from 'history'
import React, { useState, useCallback } from 'react'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import { ThemeProps } from '../../../../../shared/src/theme'
import { ExternalChangesetFields } from '../../../graphql-operations'
import { RepoSpec, RevisionSpec, FileSpec, ResolvedRevisionSpec } from '../../../../../shared/src/util/url'
import { HoverMerged } from '../../../../../shared/src/api/client/types/hover'
import { ActionItemAction } from '../../../../../shared/src/actions/ActionItem'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { ChangesetCheckStatusCell } from '../detail/changesets/ChangesetCheckStatusCell'
import { ChangesetReviewStatusCell } from '../detail/changesets/ChangesetReviewStatusCell'
import { DiffStat } from '../../../components/diff/DiffStat'
import { ErrorAlert } from '../../../components/alerts'
import { ChangesetFileDiff } from '../detail/changesets/ChangesetFileDiff'
import { queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs } from '../detail/backend'
import { ChangesetCloseActionClose, ChangesetCloseActionKept } from './ChangesetCloseAction'
import { ExternalChangesetInfoCell } from '../detail/changesets/ExternalChangesetInfoCell'

export interface ExternalChangesetCloseNodeProps extends ThemeProps {
    node: ExternalChangesetFields
    willClose: boolean
    viewerCanAdminister: boolean
    campaignUpdates?: Pick<Observer<void>, 'next'>
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
    campaignUpdates,
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
            {willClose ? <ChangesetCloseActionClose /> : <ChangesetCloseActionKept />}
            <ExternalChangesetInfoCell
                node={node}
                viewerCanAdminister={viewerCanAdminister}
                campaignUpdates={campaignUpdates}
            />
            <span>{node.checkState && <ChangesetCheckStatusCell checkState={node.checkState} />}</span>
            <span>{node.reviewState && <ChangesetReviewStatusCell reviewState={node.reviewState} />}</span>
            <div className="external-changeset-close-node__diffstat">
                {node.diffStat && <DiffStat {...node.diffStat} expandedCounts={true} />}
            </div>
            {isExpanded && (
                <div className="external-changeset-close-node__expanded-section">
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
                    />
                </div>
            )}
        </>
    )
}
