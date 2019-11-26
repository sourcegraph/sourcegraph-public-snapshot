import * as H from 'history'
import {
    IExternalChangeset,
    IChangesetPlan,
    ChangesetReviewState,
    IFileDiff,
    IPreviewFileDiff,
    ChangesetState,
} from '../../../../../../shared/src/graphql/schema'
import React from 'react'
import SourcePullIcon from 'mdi-react/SourcePullIcon'
import {
    changesetStatusColorClasses,
    changesetReviewStateColors,
    changesetReviewStateIcons,
    changesetStageLabels,
} from './presentation'
import { Link } from '../../../../../../shared/src/components/Link'
import { LinkOrSpan } from '../../../../../../shared/src/components/LinkOrSpan'
import { ThemeProps } from '../../../../../../shared/src/theme'
import { Collapsible } from '../../../../components/Collapsible'
import { DiffStat } from '../../../../components/diff/DiffStat'
import { FileDiffNode } from '../../../../components/diff/FileDiffNode'

export interface ChangesetNodeProps extends ThemeProps {
    node: IExternalChangeset | IChangesetPlan
    history: H.History
    location: H.Location
}

export const ChangesetNode: React.FunctionComponent<ChangesetNodeProps> = ({
    node,
    isLightTheme,
    history,
    location,
}) => {
    const fileDiffs = node.diff?.fileDiffs
    const fileDiffNodes: (IFileDiff | IPreviewFileDiff)[] | undefined = fileDiffs ? fileDiffs.nodes : undefined
    const ReviewStateIcon =
        node.__typename === 'ExternalChangeset'
            ? changesetReviewStateIcons[node.reviewState]
            : changesetReviewStateIcons[ChangesetReviewState.PENDING]
    return (
        <li className="list-group-item">
            <Collapsible
                title={
                    <div className="d-flex pl-1 align-items-center">
                        <div className="flex-shrink-0 flex-grow-0 m-1">
                            <SourcePullIcon
                                className={
                                    node.__typename === 'ExternalChangeset'
                                        ? `text-${changesetStatusColorClasses[node.state]}`
                                        : `text-${changesetStatusColorClasses[ChangesetState.OPEN]}`
                                }
                                data-tooltip={
                                    node.__typename === 'ExternalChangeset'
                                        ? changesetStageLabels[node.state]
                                        : changesetStageLabels[ChangesetState.OPEN]
                                }
                            />
                        </div>
                        {node.__typename === 'ExternalChangeset' && (
                            <div className="flex-shrink-0 flex-grow-0 m-1">
                                <ReviewStateIcon
                                    className={`text-${changesetReviewStateColors[node.reviewState]}`}
                                    data-tooltip={changesetStageLabels[node.reviewState]}
                                />
                            </div>
                        )}
                        <div className="flex-fill overflow-hidden m-1">
                            <h4 className="m-0">
                                <Link
                                    to={node.repository.url}
                                    className="text-muted"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                >
                                    {node.repository.name}
                                </Link>{' '}
                                {node.__typename === 'ExternalChangeset' && (
                                    <>
                                        <LinkOrSpan
                                            to={node.externalURL && node.externalURL.url}
                                            target="_blank"
                                            rel="noopener noreferrer"
                                        >
                                            {node.title}
                                        </LinkOrSpan>
                                        <div className="text-truncate w-100">{node.body}</div>
                                    </>
                                )}
                            </h4>
                            {fileDiffs && (
                                <DiffStat
                                    {...fileDiffs.diffStat}
                                    className="flex-shrink-0 flex-grow-0"
                                    expandedCounts={true}
                                ></DiffStat>
                            )}
                        </div>
                    </div>
                }
                wholeTitleClickable={false}
            >
                {fileDiffNodes &&
                    fileDiffNodes.map((fileDiffNode, i) => (
                        <FileDiffNode
                            isLightTheme={isLightTheme}
                            node={fileDiffNode}
                            lineNumbers={true}
                            location={location}
                            history={history}
                            persistLines={node.__typename === 'ExternalChangeset'}
                            key={i}
                        ></FileDiffNode>
                    ))}
            </Collapsible>
        </li>
    )
}
