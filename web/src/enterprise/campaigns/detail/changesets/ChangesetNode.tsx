import * as H from 'history'
import {
    IExternalChangeset,
    IChangesetPlan,
    ChangesetReviewState,
    IFileDiff,
    IPreviewFileDiff,
    ChangesetState,
    ChangesetCheckState,
} from '../../../../../../shared/src/graphql/schema'
import React, { useState } from 'react'
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
import { Markdown } from '../../../../../../shared/src/components/Markdown'
import { renderMarkdown } from '../../../../../../shared/src/util/markdown'
import { publishChangeset as _publishChangeset } from '../backend'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Subject } from 'rxjs'
import ErrorIcon from 'mdi-react/ErrorIcon'
import { asError } from '../../../../../../shared/src/util/errors'
import { ChangesetLabel } from './ChangesetLabel'

export interface ChangesetNodeProps extends ThemeProps {
    node: IExternalChangeset | IChangesetPlan
    campaignUpdates: Subject<void>
    history: H.History
    location: H.Location
    /** Shows the publish button for ChangesetPlans */
    enablePublishing: boolean
}

export const ChangesetNode: React.FunctionComponent<ChangesetNodeProps> = ({
    node,
    campaignUpdates,
    isLightTheme,
    history,
    location,
    enablePublishing,
}) => {
    const [isLoading, setIsLoading] = useState<boolean>(false)
    const [publishError, setPublishError] = useState<Error>()
    const publishChangeset: React.MouseEventHandler = async () => {
        try {
            setPublishError(undefined)
            setIsLoading(true)
            await _publishChangeset(node.id)
            campaignUpdates.next()
        } catch (error) {
            setPublishError(asError(error))
        } finally {
            setIsLoading(false)
        }
    }
    const fileDiffs = node.diff?.fileDiffs
    const fileDiffNodes: (IFileDiff | IPreviewFileDiff)[] | undefined = fileDiffs ? fileDiffs.nodes : undefined
    const ChangesetStateIcon =
        node.__typename === 'ExternalChangeset'
            ? changesetStateIcons[node.state]
            : changesetStateIcons[ChangesetState.OPEN]
    const ReviewStateIcon =
        node.__typename === 'ExternalChangeset'
            ? changesetReviewStateIcons[node.reviewState]
            : changesetReviewStateIcons[ChangesetReviewState.PENDING]
    const ChangesetCheckStateIcon =
        node.__typename === 'ExternalChangeset' && node.checkState
            ? changesetCheckStateIcons[node.checkState]
            : changesetCheckStateIcons[ChangesetCheckState.PENDING]
    const changesetState = node.__typename === 'ExternalChangeset' ? node.state : ChangesetState.OPEN
    const changesetNodeRow = (
        <div className="d-flex align-items-center m-1">
            <div className="flex-shrink-0 flex-grow-0 my-1 mr-1">
                <ChangesetStateIcon
                    className={`text-${changesetStatusColorClasses[changesetState]}`}
                    data-tooltip={changesetStageLabels[changesetState]}
                ></ChangesetStateIcon>
            </div>
            {node.__typename === 'ExternalChangeset' && (
                <div className="flex-shrink-0 flex-grow-0 ml-1 mr-3">
                    <ReviewStateIcon
                        className={
                            node.state === ChangesetState.DELETED
                                ? 'text-muted'
                                : `text-${changesetReviewStateColors[node.reviewState]}`
                        }
                        data-tooltip={changesetStageLabels[node.reviewState]}
                    />
                </div>
            )}
            <div className="changeset-node__content flex-fill">
                <h3 className="m-0">
                    <Link to={node.repository.url} className="text-muted" target="_blank" rel="noopener noreferrer">
                        {node.repository.name}
                    </Link>{' '}
                    {node.__typename === 'ChangesetPlan' && enablePublishing && (
                        <span className="badge badge-light">{node.publicationEnqueued ? 'Publishing' : 'Draft'}</span>
                    )}
                    <span className="mx-1"></span>{' '}
                    {node.__typename === 'ExternalChangeset' && (
                        <>
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
                                {node.title}
                            </LinkOrSpan>
                            {node.labels.length > 0 && (
                                <span className="ml-2">
                                    {node.labels.map(label => (
                                        <ChangesetLabel label={label} key={label.text} />
                                    ))}
                                </span>
                            )}
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
            <span className="ml-1 changeset-node__check-status d-flex justify-content-end">
                {node.__typename === 'ExternalChangeset' && node.checkState && (
                    <ChangesetCheckStateIcon
                        className={changesetCheckStateColors[node.checkState]}
                        data-tooltip={changesetCheckStateTooltips[node.checkState]}
                    />
                )}
            </span>
            {enablePublishing && node.__typename === 'ChangesetPlan' && !node.publicationEnqueued && (
                <>
                    {publishError && <ErrorIcon data-tooltip={publishError.message} className="ml-2" />}
                    <button
                        type="button"
                        className="flex-shrink-0 flex-grow-0 btn btn-sm btn-secondary ml-2"
                        disabled={isLoading}
                        onClick={publishChangeset}
                    >
                        {isLoading && <LoadingSpinner className="mr-1 icon-inline"></LoadingSpinner>} Publish
                    </button>
                </>
            )}
        </div>
    )
    return (
        <li className="list-group-item e2e-changeset-node">
            {fileDiffNodes ? (
                <Collapsible
                    titleClassName="changeset-node__content flex-fill"
                    title={changesetNodeRow}
                    wholeTitleClickable={false}
                >
                    {fileDiffNodes.map((fileDiffNode, i) => (
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
            ) : (
                <div className="changeset-node__content changeset-node__content--no-collapse flex-fill">
                    {changesetNodeRow}
                </div>
            )}
        </li>
    )
}
