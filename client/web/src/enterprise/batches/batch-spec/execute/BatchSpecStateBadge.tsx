import React from 'react'

import { Badge, BadgeVariantType } from '@sourcegraph/wildcard'

import { BatchSpecState } from '../../../../graphql-operations'

export interface BatchSpecStateBadgeProps {
    className?: string
    state: BatchSpecState
}

const getProps = (state: BatchSpecState): [variant: BadgeVariantType, tooltip: string] => {
    switch (state) {
        case BatchSpecState.PENDING:
            return ['secondary', 'Not started']
        case BatchSpecState.QUEUED:
            return ['secondary', 'Waiting for executor']
        case BatchSpecState.PROCESSING:
            return ['secondary', 'Currently executing']
        case BatchSpecState.CANCELED:
            return ['secondary', 'Execution has been canceled']
        case BatchSpecState.CANCELING:
            return ['secondary', 'Canceling execution']
        case BatchSpecState.FAILED:
            return ['danger', 'Execution failed']
        case BatchSpecState.COMPLETED:
            return ['success', 'Execution finished successfully']
    }
}

export const BatchSpecStateBadge: React.FunctionComponent<React.PropsWithChildren<BatchSpecStateBadgeProps>> = ({
    className,
    state,
}) => {
    const [variant, tooltip] = getProps(state)

    return (
        <Badge className={className} variant={variant} tooltip={tooltip}>
            {state}
        </Badge>
    )
}
