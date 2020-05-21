import * as H from 'history'
import {
    IExternalChangeset,
    ChangesetState,
    ChangesetCheckState,
    GitRevSpec,
    IRepositoryComparison,
} from '../../../../../../shared/src/graphql/schema'
import React, { useCallback, useMemo, useState } from 'react'
import {
    changesetReviewStateColors,
    changesetReviewStateIcons,
    changesetStageLabels,
    changesetStatusColorClasses,
    changesetStateIcons,
    changesetCheckStateIcons,
    changesetCheckStateColors,
    changesetCheckStateTooltips,
} from './presentation'
import { Link } from '../../../../../../shared/src/components/Link'
import { LinkOrSpan } from '../../../../../../shared/src/components/LinkOrSpan'
import { ThemeProps } from '../../../../../../shared/src/theme'
import { Collapsible } from '../../../../components/Collapsible'
import { DiffStat } from '../../../../components/diff/DiffStat'
import { FileDiffNode } from '../../../../components/diff/FileDiffNode'
import { publishChangeset as _publishChangeset, queryExternalChangesetWithFileDiffs } from '../backend'
import { Observer } from 'rxjs'
import { ChangesetLabel } from './ChangesetLabel'
import classNames from 'classnames'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { Hoverifier } from '@sourcegraph/codeintellify'
import { RepoSpec, RevSpec, FileSpec, ResolvedRevSpec } from '../../../../../../shared/src/util/url'
import { HoverMerged } from '../../../../../../shared/src/api/client/types/hover'
import { ActionItemAction } from '../../../../../../shared/src/actions/ActionItem'
import { FileDiffConnection } from '../../../../components/diff/FileDiffConnection'
import { FilteredConnectionQueryArgs } from '../../../../components/FilteredConnection'
import { ChangesetLastSynced } from './ChangesetLastSynced'
import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'
import { tap, map } from 'rxjs/operators'

export interface ChangesetNodeProps extends ThemeProps {
    node: IExternalChangeset
    campaignUpdates?: Pick<Observer<void>, 'next'>
    history: H.History
    location: H.Location
    extensionInfo?: {
        hoverifier: Hoverifier<RepoSpec & RevSpec & FileSpec & ResolvedRevSpec, HoverMerged, ActionItemAction>
    } & ExtensionsControllerProps
}

export const ChangesetNode: React.FunctionComponent<ChangesetNodeProps> = ({
    node,
    campaignUpdates,
    isLightTheme,
    history,
    location,
    extensionInfo,
}) => {
    const ChangesetStateIcon = changesetStateIcons[node.state]
    const ReviewStateIcon = changesetReviewStateIcons[node.reviewState]
    const ChangesetCheckStateIcon = node.checkState
        ? changesetCheckStateIcons[node.checkState]
        : changesetCheckStateIcons[ChangesetCheckState.PENDING]
    const changesetState = node.state

    const changesetNodeRow = (
        <div className="d-flex align-items-start m-1 ml-2">
            <div className="changeset-node__content flex-fill">
                <div className="d-flex flex-column">
                    <div className="m-0 mb-2">
                        <h3 className="m-0 d-inline">
                            <ChangesetStateIcon
                                className={classNames(
                                    'mr-1 icon-inline',
                                    `text-${changesetStatusColorClasses[changesetState]}`
                                )}
                                data-tooltip={changesetStageLabels[changesetState]}
                            />
                            <LinkOrSpan
                                /* Deleted changesets most likely don't exist on the codehost anymore and would return 404 pages */
                                to={
                                    node.externalURL && node.state !== ChangesetState.DELETED
                                        ? node.externalURL.url
                                        : undefined
                                }
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                {node.title} (#{node.externalID}){' '}
                                {node.externalURL && node.state !== ChangesetState.DELETED && (
                                    <ExternalLinkIcon size="1rem" />
                                )}
                            </LinkOrSpan>
                        </h3>
                        {node.checkState && (
                            <small>
                                <ChangesetCheckStateIcon
                                    className={classNames(
                                        'ml-1 changeset-node__check-state',
                                        changesetCheckStateColors[node.checkState]
                                    )}
                                    data-tooltip={changesetCheckStateTooltips[node.checkState]}
                                />
                            </small>
                        )}
                        {node.labels.length > 0 && (
                            <span className="ml-2">
                                {node.labels.map(label => (
                                    <ChangesetLabel label={label} key={label.text} />
                                ))}
                            </span>
                        )}
                    </div>
                    <div>
                        <strong>
                            <Link to={node.repository.url} target="_blank" rel="noopener noreferrer">
                                {node.repository.name}
                            </Link>
                        </strong>
                        <ChangesetLastSynced changeset={node} campaignUpdates={campaignUpdates} />
                    </div>
                </div>
            </div>
            <div className="flex-shrink-0 flex-grow-0 ml-1 align-items-end">
                {node.diff?.fileDiffs && <DiffStat {...node.diff.fileDiffs.diffStat} expandedCounts={true} />}
            </div>
            <div className="flex-shrink-0 flex-grow-0 ml-1 align-items-end">
                <ReviewStateIcon
                    className={
                        node.state === ChangesetState.DELETED
                            ? 'text-muted'
                            : `text-${changesetReviewStateColors[node.reviewState]}`
                    }
                    data-tooltip={changesetStageLabels[node.reviewState]}
                />
            </div>
        </div>
    )

    const [range, setRange] = useState<IRepositoryComparison['range']>()

    /** Fetches the file diffs for the changeset */
    const queryFileDiffs = useCallback(
        (args: FilteredConnectionQueryArgs) =>
            queryExternalChangesetWithFileDiffs(node.id, { ...args, isLightTheme }).pipe(
                map(changeset => {
                    if (!changeset.diff) {
                        throw new Error('The given changeset has no diff')
                    }
                    return changeset.diff
                }),
                tap(diff => {
                    setRange(diff.range)
                }),
                map(diff => diff.fileDiffs)
            ),
        [node.id, isLightTheme]
    )

    const hydratedExtensionInfo = useMemo(() => {
        if (!extensionInfo || !range) {
            return
        }
        const baseRev = commitOIDForGitRev(range.base)
        const headRev = commitOIDForGitRev(range.head)
        return {
            ...extensionInfo,
            head: {
                commitID: headRev,
                repoID: node.repository.id,
                repoName: node.repository.name,
                rev: headRev,
            },
            base: {
                commitID: baseRev,
                repoID: node.repository.id,
                repoName: node.repository.name,
                rev: baseRev,
            },
        }
    }, [extensionInfo, range, node.repository.id, node.repository.name])

    return (
        <li className="list-group-item e2e-changeset-node">
            {node.diff?.fileDiffs ? (
                <Collapsible
                    titleClassName="changeset-node__content flex-fill"
                    expandedButtonClassName="mb-3"
                    title={changesetNodeRow}
                    wholeTitleClickable={false}
                >
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
                </Collapsible>
            ) : (
                <div className="changeset-node__content changeset-node__content--no-collapse flex-fill">
                    {changesetNodeRow}
                </div>
            )}
        </li>
    )
}

function commitOIDForGitRev(rev: GitRevSpec): string {
    switch (rev.__typename) {
        case 'GitObject':
            return rev.oid
        case 'GitRef':
            return rev.target.oid
        case 'GitRevSpecExpr':
            if (!rev.object) {
                throw new Error('Could not resolve commit for revision')
            }
            return rev.object.oid
    }
}
