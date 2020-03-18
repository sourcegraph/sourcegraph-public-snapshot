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
import Octicon, { Diff } from '@primer/octicons-react'
import React, { useState, useEffect } from 'react'
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
import { publishChangeset as _publishChangeset, syncChangeset } from '../backend'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Subject } from 'rxjs'
import ErrorIcon from 'mdi-react/ErrorIcon'
import { asError } from '../../../../../../shared/src/util/errors'
import { ChangesetLabel } from './ChangesetLabel'
import classNames from 'classnames'
import SyncIcon from 'mdi-react/SyncIcon'
import { parseISO, formatDistance } from 'date-fns'

export interface ChangesetNodeProps extends ThemeProps {
    node: IExternalChangeset | IChangesetPlan
    campaignUpdates?: Subject<void>
    history: H.History
    location: H.Location
    /** Shows the publish button for ChangesetPlans */
    enablePublishing: boolean

    /** For testing only */
    _now?: Date
}

export const ChangesetNode: React.FunctionComponent<ChangesetNodeProps> = ({
    node,
    campaignUpdates,
    isLightTheme,
    history,
    location,
    enablePublishing,
    _now,
}) => {
    const [lastUpdatedAt, setLastUpdatedAt] = useState<string | null>(null)
    const [isPublishing, setIsPublishing] = useState<boolean>()
    const publicationEnqueued = node.__typename === 'ChangesetPlan' && node.publicationEnqueued
    useEffect(() => {
        setIsPublishing(publicationEnqueued)
    }, [publicationEnqueued])
    const nodeUpdatedAt = node.__typename === 'ExternalChangeset' && node.updatedAt
    const lastUpdatedAtChanged = lastUpdatedAt && nodeUpdatedAt !== lastUpdatedAt
    useEffect(() => {
        if (lastUpdatedAtChanged && nodeUpdatedAt) {
            if (campaignUpdates) {
                campaignUpdates.next()
            }
            setLastUpdatedAt(null)
        }
    }, [campaignUpdates, lastUpdatedAtChanged, nodeUpdatedAt])
    const [publishError, setPublishError] = useState<Error>()
    const publishChangeset: React.MouseEventHandler = async () => {
        try {
            setPublishError(undefined)
            setIsPublishing(true)
            await _publishChangeset(node.id)
            if (campaignUpdates) {
                campaignUpdates.next()
            }
        } catch (error) {
            setPublishError(asError(error))
        } finally {
            setIsPublishing(false)
        }
    }
    const enqueueChangeset: React.MouseEventHandler = async () => {
        // already enqueued
        if (lastUpdatedAt) {
            return
        }
        if (node.__typename === 'ExternalChangeset') {
            setLastUpdatedAt(node.updatedAt)
            await syncChangeset(node.id)
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

    const UpdateLoaderIcon =
        node.__typename === 'ExternalChangeset' && node.updatedAt !== lastUpdatedAt ? SyncIcon : LoadingSpinner

    const changesetNodeRow = (
        <div className="d-flex align-items-start m-1 ml-2">
            <div className="changeset-node__content flex-fill">
                <div className="d-flex flex-column">
                    {node.__typename === 'ExternalChangeset' && (
                        <div className="m-0 mb-2">
                            <ChangesetStateIcon
                                className={classNames(
                                    'mr-1 icon-inline',
                                    `text-${changesetStatusColorClasses[changesetState]}`
                                )}
                                data-tooltip={changesetStageLabels[changesetState]}
                            />
                            <h3 className="m-0 d-inline">
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
                                    {node.title} (#{node.externalID})
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
                    )}
                    <div>
                        {node.__typename === 'ChangesetPlan' && <Octicon icon={Diff} className="icon-inline mr-2" />}
                        <strong>
                            <Link
                                to={node.repository.url}
                                className={classNames(node.__typename === 'ChangesetPlan' && 'text-muted')}
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                {node.repository.name}
                            </Link>
                        </strong>
                        {node.__typename === 'ExternalChangeset' && (
                            <small className="text-muted ml-2">
                                Last synced {formatDistance(parseISO(node.updatedAt), _now ?? new Date())} ago.{' '}
                                <span
                                    data-tooltip={
                                        node.updatedAt === lastUpdatedAt
                                            ? 'Currently refreshing'
                                            : 'Click to prioritize refresh'
                                    }
                                >
                                    <UpdateLoaderIcon
                                        className={classNames('icon-inline', !lastUpdatedAt && 'cursor-pointer')}
                                        onClick={enqueueChangeset}
                                    />
                                </span>
                            </small>
                        )}
                    </div>
                </div>
            </div>
            <div className="flex-shrink-0 flex-grow-0 ml-1 align-items-end">
                {fileDiffs && <DiffStat {...fileDiffs.diffStat} expandedCounts={true} />}
            </div>
            <div className="flex-shrink-0 flex-grow-0 ml-1 align-items-end">
                {node.__typename === 'ExternalChangeset' && (
                    <ReviewStateIcon
                        className={
                            node.state === ChangesetState.DELETED
                                ? 'text-muted'
                                : `text-${changesetReviewStateColors[node.reviewState]}`
                        }
                        data-tooltip={changesetStageLabels[node.reviewState]}
                    />
                )}
            </div>
            {enablePublishing && node.__typename === 'ChangesetPlan' && (
                <>
                    {publishError && <ErrorIcon data-tooltip={publishError.message} className="ml-2" />}
                    <button
                        type="button"
                        className="flex-shrink-0 flex-grow-0 btn btn-sm btn-secondary ml-2"
                        disabled={isPublishing}
                        onClick={publishChangeset}
                    >
                        {isPublishing && <LoadingSpinner className="mr-1 icon-inline" />}{' '}
                        {isPublishing ? 'Publishing' : 'Publish'}
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
                    {fileDiffNodes.map(fileDiffNode => (
                        <FileDiffNode
                            isLightTheme={isLightTheme}
                            node={fileDiffNode}
                            lineNumbers={true}
                            location={location}
                            history={history}
                            persistLines={node.__typename === 'ExternalChangeset'}
                            key={fileDiffNode.internalID}
                            className="mb-1"
                        />
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
