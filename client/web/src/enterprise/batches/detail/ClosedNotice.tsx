import React from 'react'

import { Alert } from '@sourcegraph/wildcard'

import type { BatchChangeFields } from '../../../graphql-operations'

interface ClosedNoticeProps {
    closedAt: BatchChangeFields['closedAt']
    className?: string
}

export const ClosedNotice: React.FunctionComponent<React.PropsWithChildren<ClosedNoticeProps>> = ({
    closedAt,
    className,
}) => {
    if (!closedAt) {
        return null
    }

    return (
        <Alert className={className} variant="info">
            Information on this page may be out of date because changesets that only exist in closed batch changes are
            not synced with the code host.
        </Alert>
    )
}
