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
import { Markdown } from '../../../../../../shared/src/components/Markdown'
import { renderMarkdown } from '../../../../../../shared/src/util/markdown'

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
                titleClassName="campaign-node__content flex-fill"
                title={
                    <div className="d-flex align-items-center m-1">
                        <div className="flex-shrink-0 flex-grow-0 my-1 mr-1">
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
                            <div className="flex-shrink-0 flex-grow-0 ml-1 mr-3">
                                <ReviewStateIcon
                                    className={`text-${changesetReviewStateColors[node.reviewState]}`}
                                    data-tooltip={changesetStageLabels[node.reviewState]}
                                />
                            </div>
                        )}
                        <div className="campaign-node__content flex-fill">
                            <h3 className="m-0">
                                <Link
                                    to={node.repository.url}
                                    className="text-muted"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                >
                                    {node.repository.name}
                                </Link>{' '}
                                <span className="mx-1"></span>{' '}
                                {node.__typename === 'ExternalChangeset' && (
                                    <>
                                        <LinkOrSpan
                                            to={node.externalURL && node.externalURL.url}
                                            target="_blank"
                                            rel="noopener noreferrer"
                                        >
                                            {node.title}
                                        </LinkOrSpan>
                                    </>
                                )}
                            </h3>
                            {node.__typename === 'ExternalChangeset' && (
                                <Markdown
                                    className="text-truncate"
                                    dangerousInnerHTML={renderMarkdown(node.body, { plainText: true })}
                                ></Markdown>
                            )}
                        </div>
                        {fileDiffs && (
                            <span className="flex-shrink-0 flex-grow-0">
                                <DiffStat {...fileDiffs.diffStat} expandedCounts={true}></DiffStat>
                            </span>
                        )}
                    </div>
                }
                wholeTitleClickable={false}
            >
                {fileDiffNodes?.map((fileDiffNode, i) => (
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
