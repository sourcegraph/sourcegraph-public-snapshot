import React, { useState, useEffect, useCallback } from 'react'

import { mdiAlertCircle, mdiSync, mdiInformationOutline } from '@mdi/js'
import { formatDistance, isBefore, parseISO } from 'date-fns'

import { isErrorLike } from '@sourcegraph/common'
import { LoadingSpinner, Icon, Tooltip, Button } from '@sourcegraph/wildcard'

import type { ExternalChangesetFields, HiddenExternalChangesetFields } from '../../../../graphql-operations'
import { syncChangeset } from '../backend'

import styles from './ChangesetLastSynced.module.scss'

interface Props {
    changeset:
        | Pick<HiddenExternalChangesetFields, 'id' | 'nextSyncAt' | 'updatedAt' | '__typename'>
        | Pick<ExternalChangesetFields, 'id' | 'nextSyncAt' | 'updatedAt' | '__typename' | 'syncerError'>
    viewerCanAdminister: boolean
    /** For testing purposes only */
    _now?: Date
}

export const ChangesetLastSynced: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    changeset,
    viewerCanAdminister,
    _now,
}) => {
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

    return (
        <small className="text-muted">
            {changeset.__typename === 'ExternalChangeset' && changeset.syncerError ? (
                <Tooltip content="Expand to see details.">
                    <span>
                        <Icon aria-hidden={true} className="text-danger" svgPath={mdiAlertCircle} /> Syncing from code
                        host failed.
                    </span>
                </Tooltip>
            ) : (
                <>{`Last synced ${formatDistance(parseISO(changeset.updatedAt), _now ?? new Date())} ago.`}</>
            )}{' '}
            {isErrorLike(lastUpdatedAt) && (
                <Tooltip content={lastUpdatedAt.message}>
                    <Icon aria-label={lastUpdatedAt.message} className="ml-2 small" svgPath={mdiAlertCircle} />
                </Tooltip>
            )}
            <Tooltip content={tooltipText}>
                <span className={styles.updateLoaderWrapper}>
                    <UpdateLoaderIcon
                        changesetUpdatedAt={changeset.updatedAt}
                        lastUpdatedAt={lastUpdatedAt}
                        onEnqueueChangeset={enqueueChangeset}
                        viewerCanAdminister={viewerCanAdminister}
                    />
                </span>
            </Tooltip>
        </small>
    )
}

const UpdateLoaderIcon: React.FunctionComponent<
    React.PropsWithChildren<{
        lastUpdatedAt: string | Error | null
        changesetUpdatedAt: string
        viewerCanAdminister: boolean
        onEnqueueChangeset: React.MouseEventHandler
    }>
> = ({ lastUpdatedAt, changesetUpdatedAt, onEnqueueChangeset, viewerCanAdminister }) => {
    if (typeof lastUpdatedAt === 'string' && changesetUpdatedAt === lastUpdatedAt) {
        return <LoadingSpinner inline={true} />
    }

    if (viewerCanAdminister) {
        return (
            <Button aria-label="Refresh" variant="icon" className="d-inline" onClick={onEnqueueChangeset}>
                <Icon className="text-muted" svgPath={mdiSync} aria-hidden={true} />
            </Button>
        )
    }

    return <Icon aria-hidden={true} svgPath={mdiInformationOutline} />
}
