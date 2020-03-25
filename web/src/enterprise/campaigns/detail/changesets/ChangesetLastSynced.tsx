import React, { useState, useEffect } from 'react'
import { IExternalChangeset } from '../../../../../../shared/src/graphql/schema'
import classNames from 'classnames'
import { formatDistance, parseISO } from 'date-fns'
import { syncChangeset } from '../backend'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import SyncIcon from 'mdi-react/SyncIcon'
import { Observer } from 'rxjs'

interface Props {
    changeset: Pick<IExternalChangeset, 'id' | 'updatedAt'>
    campaignUpdates?: Pick<Observer<void>, 'next'>

    /** For testing purposes only */
    _now?: Date
}

export const ChangesetLastSynced: React.FunctionComponent<Props> = ({ changeset, campaignUpdates, _now }) => {
    // initially, the changeset was never last updated
    const [lastUpdatedAt, setLastUpdatedAt] = useState<string | null>(null)
    // .. if it was, and the changesets current updatedAt doesn't match the previous updated at, we know that it has been synced
    const lastUpdatedAtChanged = lastUpdatedAt && changeset.updatedAt !== lastUpdatedAt
    useEffect(() => {
        if (lastUpdatedAtChanged) {
            if (campaignUpdates) {
                campaignUpdates.next()
            }
            setLastUpdatedAt(null)
        }
    }, [campaignUpdates, lastUpdatedAtChanged, changeset.updatedAt])

    const enqueueChangeset: React.MouseEventHandler = async () => {
        // already enqueued
        if (lastUpdatedAt) {
            return
        }
        setLastUpdatedAt(changeset.updatedAt)
        await syncChangeset(changeset.id)
    }

    const UpdateLoaderIcon = changeset.updatedAt !== lastUpdatedAt ? SyncIcon : LoadingSpinner

    return (
        <small className="text-muted ml-2">
            Last synced {formatDistance(parseISO(changeset.updatedAt), _now ?? new Date())} ago.{' '}
            <span
                data-tooltip={
                    changeset.updatedAt === lastUpdatedAt ? 'Currently refreshing' : 'Click to prioritize refresh'
                }
            >
                <UpdateLoaderIcon
                    className={classNames('icon-inline', !lastUpdatedAt && 'cursor-pointer')}
                    onClick={enqueueChangeset}
                />
            </span>
        </small>
    )
}
