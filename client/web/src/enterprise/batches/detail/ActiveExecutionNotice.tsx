import React, { useMemo } from 'react'

import { pluralize } from '@sourcegraph/common'
import { AlertLink, Alert } from '@sourcegraph/wildcard'

import { BatchSpecState } from '../../../graphql-operations'

interface ActiveExecutionNoticeProps {
    batchSpecs: { state: BatchSpecState }[]
    batchChangeURL: string
    className?: string
}

export const ActiveExecutionNotice: React.FunctionComponent<React.PropsWithChildren<ActiveExecutionNoticeProps>> = ({
    batchSpecs,
    batchChangeURL,
    className,
}) => {
    const numberExecuting = useMemo(
        () =>
            batchSpecs.filter(({ state }) => state === BatchSpecState.PROCESSING || state === BatchSpecState.QUEUED)
                .length,
        [batchSpecs]
    )

    if (!numberExecuting) {
        return null
    }

    return (
        <Alert className={className} variant="waiting">
            There {pluralize('is', numberExecuting, 'are')} currently {numberExecuting} batch{' '}
            {pluralize('spec', numberExecuting)} <AlertLink to={`${batchChangeURL}/executions`}>executing</AlertLink>.
        </Alert>
    )
}
