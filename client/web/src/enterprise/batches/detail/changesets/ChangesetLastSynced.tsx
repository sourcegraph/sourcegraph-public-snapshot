import classNames from 'classnames'
import { formatDistance, isBefore, parseISO } from 'date-fns'
import ErrorIcon from 'mdi-react/ErrorIcon'
import InfoCircleOutlineIcon from 'mdi-react/InfoCircleOutlineIcon'
import SyncIcon from 'mdi-react/SyncIcon'
import React, { useState, useEffect, useCallback } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { ExternalChangesetFields, HiddenExternalChangesetFields } from '../../../../graphql-operations'
import { syncChangeset } from '../backend'

interface Props {
    changeset:
        | Pick<HiddenExternalChangesetFields, 'id' | 'nextSyncAt' | 'updatedAt' | '__typename'>
        | Pick<ExternalChangesetFields, 'id' | 'nextSyncAt' | 'updatedAt' | '__typename' | 'syncerError'>
    viewerCanAdminister: boolean
    /** For testing purposes only */
    _now?: Date
}

export const ChangesetLastSynced: React.FunctionComponent<Props> = ({ changeset, viewerCanAdminister, _now }) => {
    // initially, the changeset was never last updated
    const [lastUpdatedAt, setLastUpdatedAt] = useState<string | Error | null>(null)
    // .. if it was, and the changesets current updatedAt doesn't match the previous updated at, we know that it has been synced
    const lastUpdatedAtChanged = lastUpdatedAt && !isErrorLike(lastUpdatedAt) && changeset.updatedAt !== lastUpdatedAt
    useEffect(() => {
        if (lastUpdatedAtChanged) {
            setLastUpdatedAt(null)
        }
    }, [lastUpdatedAtChanged, changeset.updatedAt])
    const enqueueChangeset = useCallback<React.MouseEventHandler>(async () => {
        if (!viewerCanAdminister) {
            return
        }
        // already enqueued
        if (typeof lastUpdatedAt === 'string') {
            return
        }
        setLastUpdatedAt(changeset.updatedAt)
        try {
            await syncChangeset(changeset.id)
        } catch (error) {
            setLastUpdatedAt(error)
        }
    }, [changeset.id, changeset.updatedAt, lastUpdatedAt, viewerCanAdminister])

    let tooltipText = ''
    if (changeset.updatedAt === lastUpdatedAt) {
        tooltipText = 'Currently refreshing'
    } else {
        // If no nextSyncAt is set, the syncer won't pick it up for now.
        if (!changeset.nextSyncAt) {
            tooltipText = 'Not scheduled for syncing.'
            // If the nextSyncAt date is in the past, the syncer couldn't catch up
            // but it will be synced as soon as it can.
        } else if (isBefore(parseISO(changeset.nextSyncAt), _now ?? new Date())) {
            tooltipText = 'Next refresh soon.'
            // Else, nextSyncAt is set and in the future, so we can tell the approximate
            // time when the changeset will be synced again.
        } else {
            tooltipText = `Next refresh in ${formatDistance(parseISO(changeset.nextSyncAt), _now ?? new Date())}.`
        }
        if (viewerCanAdminister) {
            tooltipText += ' Click to prioritize refresh'
        }
    }

    const UpdateLoaderIcon =
        typeof lastUpdatedAt === 'string' && changeset.updatedAt === lastUpdatedAt
            ? LoadingSpinner
            : viewerCanAdminister
            ? SyncIcon
            : InfoCircleOutlineIcon

    return (
        <small className="text-muted">
            {changeset.__typename === 'ExternalChangeset' && changeset.syncerError ? (
                <span data-tooltip="Expand to see details.">
                    <ErrorIcon className="icon-inline text-danger" /> Syncing from code host failed.
                </span>
            ) : (
                <>Last synced {formatDistance(parseISO(changeset.updatedAt), _now ?? new Date())} ago.</>
            )}{' '}
            {isErrorLike(lastUpdatedAt) && (
                <ErrorIcon data-tooltip={lastUpdatedAt.message} className="ml-2 icon-inline small" />
            )}
            <span data-tooltip={tooltipText}>
                <UpdateLoaderIcon
                    className={classNames(
                        'icon-inline',
                        typeof lastUpdatedAt !== 'string' && viewerCanAdminister && 'cursor-pointer'
                    )}
                    onClick={enqueueChangeset}
                />
            </span>
        </small>
    )
}
