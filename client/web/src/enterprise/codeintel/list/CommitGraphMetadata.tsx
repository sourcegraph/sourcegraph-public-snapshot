import classNames from 'classnames'
import React, { FunctionComponent } from 'react'

import { Timestamp } from '@sourcegraph/web/src/components/time/Timestamp'

export interface CommitGraphMetadataProps {
    stale: boolean
    updatedAt: Date | null
    className?: string
    now?: () => Date
}

export const CommitGraphMetadata: FunctionComponent<CommitGraphMetadataProps> = ({
    stale,
    updatedAt,
    className,
    now,
}) => (
    <>
        <div className={classNames('alert', stale ? 'alert-primary' : 'alert-success', className)}>
            {stale ? <StaleRepository /> : <FreshRepository />}{' '}
            {updatedAt && <LastUpdated updatedAt={updatedAt} now={now} />}
        </div>
    </>
)

const FreshRepository: FunctionComponent<{}> = () => <>Repository commit graph is currently up to date.</>

const StaleRepository: FunctionComponent<{}> = () => (
    <>
        Repository commit graph is currently stale and is queued to be refreshed. Refreshing the commit graph updates
        which uploads are visible from which commits.
    </>
)

interface LastUpdatedProps {
    updatedAt: Date
    now?: () => Date
}

const LastUpdated: FunctionComponent<LastUpdatedProps> = ({ updatedAt, now }) => (
    <>
        Last refreshed <Timestamp date={updatedAt} now={now} />.
    </>
)
