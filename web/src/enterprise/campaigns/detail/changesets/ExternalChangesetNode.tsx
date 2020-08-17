import { ThemeProps } from '../../../../../../shared/src/theme'
import { Observer } from 'rxjs'
import { Hoverifier } from '@sourcegraph/codeintellify'
import { RepoSpec, RevisionSpec, FileSpec, ResolvedRevisionSpec } from '../../../../../../shared/src/util/url'
import { HoverMerged } from '../../../../../../shared/src/api/client/types/hover'
import { ActionItemAction } from '../../../../../../shared/src/actions/ActionItem'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as H from 'history'
import React, { useState, useCallback, useMemo } from 'react'
import { LinkOrSpan } from '../../../../../../shared/src/components/LinkOrSpan'
import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'
import { ChangesetLabel } from './ChangesetLabel'
import { Link } from '../../../../../../shared/src/components/Link'
import { ChangesetLastSynced } from './ChangesetLastSynced'
import { DiffStat } from '../../../../components/diff/DiffStat'
import { FilteredConnectionQueryArgs } from '../../../../components/FilteredConnection'
import { queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs } from '../backend'
import { FileDiffConnection } from '../../../../components/diff/FileDiffConnection'
import { FileDiffNode } from '../../../../components/diff/FileDiffNode'
import { tap, map } from 'rxjs/operators'
import {
    GitRefSpecFields,
    ExternalChangesetFileDiffsFields,
    ChangesetExternalState,
    ExternalChangesetFields,
    ChangesetPublicationState,
} from '../../../../graphql-operations'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import { ChangesetStatusCell } from './ChangesetStatusCell'
import { ChangesetCheckStatusCell } from './ChangesetCheckStatusCell'
import { ChangesetReviewStatusCell } from './ChangesetReviewStatusCell'
import { ErrorAlert } from '../../../../components/alerts'

export interface ExternalChangesetNodeProps extends ThemeProps {
    node: ExternalChangesetFields
    viewerCanAdminister: boolean
    campaignUpdates?: Pick<Observer<void>, 'next'>
    history: H.History
    location: H.Location
    extensionInfo?: {
        hoverifier: Hoverifier<RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec, HoverMerged, ActionItemAction>
    } & ExtensionsControllerProps
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
}

export const ExternalChangesetNode: React.FunctionComponent<ExternalChangesetNodeProps> = ({
    node,
    viewerCanAdminister,
    campaignUpdates,
    isLightTheme,
    history,
    location,
    extensionInfo,
    queryExternalChangesetWithFileDiffs = _queryExternalChangesetWithFileDiffs,
}) => {
    const [isExpanded, setIsExpanded] = useState(false)
    const toggleIsExpanded = useCallback<React.MouseEventHandler<HTMLButtonElement>>(
        event => {
            event.preventDefault()
            setIsExpanded(!isExpanded)
        },
        [isExpanded]
    )

    const [range, setRange] = useState<
        (NonNullable<ExternalChangesetFileDiffsFields['diff']> & { __typename: 'RepositoryComparison' })['range']
    >()

    /** Fetches the file diffs for the changeset */
    const queryFileDiffs = useCallback(
        (args: FilteredConnectionQueryArgs) =>
            queryExternalChangesetWithFileDiffs({
                after: args.after ?? null,
                first: args.first ?? null,
                externalChangeset: node.id,
                isLightTheme,
            }).pipe(
                map(changeset => {
                    if (!changeset.diff) {
                        throw new Error('The given changeset has no diff')
                    }
                    return changeset.diff
                }),
                tap(diff => {
                    if (diff.__typename === 'RepositoryComparison') {
                        setRange(diff.range)
                    }
                }),
                map(diff => diff.fileDiffs)
            ),
        [node.id, isLightTheme, queryExternalChangesetWithFileDiffs]
    )

    const hydratedExtensionInfo = useMemo(() => {
        if (!extensionInfo || !range) {
            return
        }
        const baseRevision = commitOIDForGitRevision(range.base)
        const headRevision = commitOIDForGitRevision(range.head)
        return {
            ...extensionInfo,
            head: {
                commitID: headRevision,
                repoID: node.repository.id,
                repoName: node.repository.name,
                revision: headRevision,
            },
            base: {
                commitID: baseRevision,
                repoID: node.repository.id,
                repoName: node.repository.name,
                revision: baseRevision,
            },
        }
    }, [extensionInfo, range, node.repository.id, node.repository.name])

    return (
        <>
            <button
                type="button"
                className="btn btn-icon"
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
            <div className="d-flex flex-column">
                <div className="m-0 mb-2">
                    <h3 className="m-0 d-inline">
                        <LinkOrSpan
                            /* Deleted changesets most likely don't exist on the codehost anymore and would return 404 pages */
                            to={
                                node.externalURL && node.externalState !== ChangesetExternalState.DELETED
                                    ? node.externalURL.url
                                    : undefined
                            }
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            {node.title}
                            {node.externalID && <>(#{node.externalID}) </>}
                            {node.externalURL && node.externalState !== ChangesetExternalState.DELETED && (
                                <>
                                    {' '}
                                    <ExternalLinkIcon size="1rem" />
                                </>
                            )}
                        </LinkOrSpan>
                    </h3>
                    {node.labels.length > 0 && (
                        <span className="ml-2">
                            {node.labels.map(label => (
                                <ChangesetLabel label={label} key={label.text} />
                            ))}
                        </span>
                    )}
                </div>
                <div>
                    <strong className="mr-2">
                        <Link to={node.repository.url} target="_blank" rel="noopener noreferrer">
                            {node.repository.name}
                        </Link>
                    </strong>
                    {node.publicationState === ChangesetPublicationState.PUBLISHED && (
                        <ChangesetLastSynced
                            changeset={node}
                            viewerCanAdminister={viewerCanAdminister}
                            campaignUpdates={campaignUpdates}
                        />
                    )}
                </div>
            </div>
            <span>{node.checkState && <ChangesetCheckStatusCell checkState={node.checkState} />}</span>
            <span>{node.reviewState && <ChangesetReviewStatusCell reviewState={node.reviewState} />}</span>
            <div className="visible-changeset-spec-node__diffstat">
                {node.diffStat && <DiffStat {...node.diffStat} expandedCounts={true} />}
            </div>
            {isExpanded && (
                <div className="visible-changeset-spec-node__expanded-section">
                    {node.error && <ErrorAlert error={node.error} history={history} />}
                    <FileDiffConnection
                        listClassName="list-group list-group-flush"
                        noun="changed file"
                        pluralNoun="changed files"
                        queryConnection={queryFileDiffs}
                        nodeComponent={FileDiffNode}
                        nodeComponentProps={{
                            history,
                            location,
                            isLightTheme,
                            persistLines: true,
                            extensionInfo: hydratedExtensionInfo,
                            lineNumbers: true,
                        }}
                        updateOnChange={node.repository.id}
                        defaultFirst={15}
                        hideSearch={true}
                        noSummaryIfAllNodesVisible={true}
                        history={history}
                        location={location}
                        useURLQuery={false}
                        cursorPaging={true}
                    />
                </div>
            )}
        </>
    )
}

function commitOIDForGitRevision(revision: GitRefSpecFields): string {
    switch (revision.__typename) {
        case 'GitObject':
            return revision.oid
        case 'GitRef':
            return revision.target.oid
        case 'GitRevSpecExpr':
            if (!revision.object) {
                throw new Error('Could not resolve commit for revision')
            }
            return revision.object.oid
    }
}
