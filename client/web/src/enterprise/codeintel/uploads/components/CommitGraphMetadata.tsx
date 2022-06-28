import { FunctionComponent } from 'react'

import { Alert } from '@sourcegraph/wildcard'

import { Timestamp } from '../../../../components/time/Timestamp'

export interface CommitGraphMetadataProps {
    stale: boolean
    updatedAt: Date | null
    className?: string
    now?: () => Date
}

export const CommitGraphMetadata: FunctionComponent<React.PropsWithChildren<CommitGraphMetadataProps>> = ({
    stale,
    updatedAt,
    className,
    now,
}) => (
    <>
        <Alert variant={stale ? 'primary' : 'success'} className={className} aria-live="off">
            {stale ? <StaleRepository /> : <FreshRepository />}{' '}
            {updatedAt && <LastUpdated updatedAt={updatedAt} now={now} />}
        </Alert>
    </>
)

const FreshRepository: FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <>Repository commit graph is currently up to date.</>
)

const StaleRepository: FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <>
        Repository commit graph is currently stale and is queued to be refreshed. Refreshing the commit graph updates
        which uploads are visible from which commits.
    </>
)

interface LastUpdatedProps {
    updatedAt: Date
    now?: () => Date
}

const LastUpdated: FunctionComponent<React.PropsWithChildren<LastUpdatedProps>> = ({ updatedAt, now }) => (
    <>
        Last refreshed <Timestamp date={updatedAt} now={now} />.
    </>
)
