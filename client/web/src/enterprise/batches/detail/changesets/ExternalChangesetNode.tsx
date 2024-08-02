import React, { useState, useCallback, useEffect } from 'react'

import { mdiChevronDown, mdiChevronRight, mdiSync } from '@mdi/js'
import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'

import { asError, isErrorLike } from '@sourcegraph/common'
import { ChangesetState } from '@sourcegraph/shared/src/graphql-operations'
import { Button, Alert, Icon, H4, Text, ErrorMessage, ErrorAlert } from '@sourcegraph/wildcard'

import { DiffStatStack } from '../../../../components/diff/DiffStat'
import { InputTooltip } from '../../../../components/InputTooltip'
import { ChangesetSpecType, type ExternalChangesetFields } from '../../../../graphql-operations'
import {
    type queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs,
    reenqueueChangeset,
} from '../backend'

import { ChangesetCheckStatusCell } from './ChangesetCheckStatusCell'
import { ChangesetFileDiff } from './ChangesetFileDiff'
import { ChangesetReviewStatusCell } from './ChangesetReviewStatusCell'
import { ChangesetStatusCell } from './ChangesetStatusCell'
import { DownloadDiffButton } from './DownloadDiffButton'
import { ExternalChangesetInfoCell } from './ExternalChangesetInfoCell'

import styles from './ExternalChangesetNode.module.scss'

export interface ExternalChangesetNodeProps {
    node: ExternalChangesetFields
    viewerCanAdminister: boolean
    selectable?: {
        onSelect: (id: string) => void
        isSelected: (id: string) => boolean
    }
    /** For testing only. */
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
    /** For testing only. */
    expandByDefault?: boolean
}

export const ExternalChangesetNode: React.FunctionComponent<React.PropsWithChildren<ExternalChangesetNodeProps>> = ({
    node: initialNode,
    viewerCanAdminister,
    selectable,
    queryExternalChangesetWithFileDiffs,
    expandByDefault,
}) => {
    const [node, setNode] = useState(initialNode)
    useEffect(() => {
        setNode(initialNode)
    }, [initialNode])
    const [isExpanded, setIsExpanded] = useState(expandByDefault ?? false)
    const toggleIsExpanded = useCallback<React.MouseEventHandler<HTMLButtonElement>>(
        event => {
            event.preventDefault()
            setIsExpanded(!isExpanded)
        },
        [isExpanded]
    )

    const selected = selectable?.isSelected(node.id)
    const toggleSelected = useCallback((): void => {
        selectable?.onSelect(node.id)
    }, [selectable, node.id])

    const tooltipLabel = viewerCanAdminister
        ? 'Click to select changeset for bulk operation'
        : 'You do not have permission to perform this operation'

    return (
        <>
            <Button
                variant="icon"
                className="test-batches-expand-changeset d-none d-sm-block"
                aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                onClick={toggleIsExpanded}
            >
                <Icon aria-hidden={true} svgPath={isExpanded ? mdiChevronDown : mdiChevronRight} />
            </Button>
            {selectable ? (
                <div className="p-2">
                    {/* eslint-disable-next-line no-restricted-syntax*/}
                    <InputTooltip
                        id={`select-changeset-${node.id}`}
                        type="checkbox"
                        checked={selected}
                        onChange={toggleSelected}
                        disabled={!viewerCanAdminister}
                        tooltip={tooltipLabel}
                        placement="right"
                        aria-label={tooltipLabel}
                    />
                </div>
            ) : (
                // 0-width empty element to allow us to keep the identical grid template of the parent
                // list, regardless of whether or not the nodes have the checkbox selector
                <span />
            )}
            <ChangesetStatusCell
                id={node.id}
                state={node.state}
                className={classNames(
                    styles.externalChangesetNodeState,
                    'p-2 align-self-stretch text-muted d-block d-sm-flex'
                )}
            />
            <ExternalChangesetInfoCell
                node={node}
                viewerCanAdminister={viewerCanAdminister}
                className={classNames(styles.externalChangesetNodeInformation, 'p-2 align-self-stretch')}
            />
            <div
                className={classNames(
                    styles.externalChangesetNodeStatuses,
                    'd-flex d-md-none justify-content-start',
                    (node.checkState || node.reviewState || node.diffStat) && 'p-2'
                )}
            >
                {node.checkState ? (
                    <ChangesetCheckStatusCell checkState={node.checkState} className="mr-3" />
                ) : (
                    <VisuallyHidden>No check state available</VisuallyHidden>
                )}
                {node.reviewState ? (
                    <ChangesetReviewStatusCell reviewState={node.reviewState} className="mr-3" />
                ) : (
                    <VisuallyHidden>No review state available</VisuallyHidden>
                )}
                {node.diffStat ? (
                    <DiffStatStack {...node.diffStat} />
                ) : (
                    <VisuallyHidden>No diff available</VisuallyHidden>
                )}
            </div>
            <span
                className={classNames(
                    'align-self-stretch d-none d-md-flex justify-content-center',
                    node.checkState && 'p-2'
                )}
            >
                {node.checkState ? (
                    <ChangesetCheckStatusCell checkState={node.checkState} />
                ) : (
                    <VisuallyHidden>No check state available</VisuallyHidden>
                )}
            </span>
            <span
                className={classNames(
                    'align-self-stretch d-none d-md-flex justify-content-center',
                    node.reviewState && 'p-2'
                )}
            >
                {node.reviewState ? (
                    <ChangesetReviewStatusCell reviewState={node.reviewState} />
                ) : (
                    <VisuallyHidden>No review state available</VisuallyHidden>
                )}
            </span>
            <div
                className={classNames(
                    'align-self-center d-none d-md-flex justify-content-center',
                    node.diffStat && 'p-2'
                )}
            >
                {node.diffStat ? (
                    <DiffStatStack {...node.diffStat} />
                ) : (
                    <VisuallyHidden>No diff available</VisuallyHidden>
                )}
            </div>
            {/* The button for expanding the information used on xs devices. */}
            <Button
                onClick={toggleIsExpanded}
                className={classNames(
                    styles.externalChangesetNodeShowDetails,
                    'd-block d-sm-none test-batches-expand-changeset'
                )}
                outline={true}
                variant="secondary"
            >
                <Icon aria-hidden={true} svgPath={isExpanded ? mdiChevronDown : mdiChevronRight} />{' '}
                {isExpanded ? 'Hide' : 'Show'} details
            </Button>
            {isExpanded && (
                <>
                    <div className={classNames(styles.externalChangesetNodeBgExpanded, 'align-self-stretch')} />
                    <div
                        className={classNames(
                            styles.externalChangesetNodeExpandedSection,
                            styles.externalChangesetNodeBgExpanded,
                            'p-2'
                        )}
                    >
                        <div className="d-flex justify-content-end">
                            {viewerCanAdminister && node.state === ChangesetState.FAILED && node.error && (
                                <RetryChangesetButton
                                    node={node}
                                    setNode={setNode}
                                    viewerCanAdminister={viewerCanAdminister}
                                />
                            )}
                            {node.currentSpec?.type === ChangesetSpecType.BRANCH && (
                                <DownloadDiffButton changesetID={node.id} />
                            )}
                        </div>
                        {node.syncerError && <SyncerError syncerError={node.syncerError} />}
                        <ChangesetError node={node} />
                        <ChangesetFileDiff
                            changesetID={node.id}
                            repositoryID={node.repository.id}
                            repositoryName={node.repository.name}
                            queryExternalChangesetWithFileDiffs={queryExternalChangesetWithFileDiffs}
                            updateOnChange={node.updatedAt}
                        />
                    </div>
                </>
            )}
        </>
    )
}

const SyncerError: React.FunctionComponent<React.PropsWithChildren<{ syncerError: string }>> = ({ syncerError }) => (
    <Alert role="alert" variant="danger">
        <H4 className={classNames(styles.alertHeading)}>
            Encountered error during last attempt to sync changeset data from code host
        </H4>
        <ErrorMessage error={syncerError} />
        <hr className="my-2" />
        <Text className="mb-0">
            <small>This might be an ephemeral error that resolves itself at the next sync.</small>
        </Text>
    </Alert>
)

const ChangesetError: React.FunctionComponent<
    React.PropsWithChildren<{
        node: ExternalChangesetFields
    }>
> = ({ node }) => {
    if (!node.error) {
        return null
    }

    if (node.state === ChangesetState.FAILED) {
        return (
            <Alert role="alert" variant="danger">
                <H4 className={classNames(styles.alertHeading)}>The changeset has failed to run the operations.</H4>
                <ErrorMessage error={node.error} />
            </Alert>
        )
    }
    return (
        <Alert role="alert" variant="warning">
            <H4 className={classNames(styles.alertHeading)}>The changeset encountered an error, but is retrying.</H4>
            <ErrorMessage error={node.error} />
        </Alert>
    )
}

const RetryChangesetButton: React.FunctionComponent<
    React.PropsWithChildren<{
        node: ExternalChangesetFields
        setNode: (node: ExternalChangesetFields) => void
        viewerCanAdminister: boolean
    }>
> = ({ node, setNode }) => {
    const [isLoading, setIsLoading] = useState<boolean | Error>(false)
    const onRetry = useCallback(async () => {
        setIsLoading(true)
        try {
            const changeset = await reenqueueChangeset(node.id)
            // If repository permissions changed in between - ignore and await fetch (at most 5s) to reflect the new state.
            if (changeset.__typename === 'ExternalChangeset') {
                setIsLoading(false)
                setNode(changeset)
            }
        } catch (error) {
            setIsLoading(asError(error))
        }
    }, [node.id, setNode])
    return (
        <>
            {isErrorLike(isLoading) && <ErrorAlert error={isLoading} prefix="Error re-enqueueing changeset" />}
            <Button className="mb-1" onClick={onRetry} disabled={isLoading === true} variant="link">
                <Icon
                    aria-hidden={true}
                    className={classNames(isLoading === true && styles.externalChangesetNodeRetrySpinning)}
                    svgPath={mdiSync}
                />{' '}
                Retry
            </Button>
        </>
    )
}
