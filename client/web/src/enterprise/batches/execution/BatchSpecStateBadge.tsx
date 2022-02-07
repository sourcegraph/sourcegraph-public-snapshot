import React from 'react'

import { Badge } from '@sourcegraph/wildcard'

import { BatchSpecState } from '../../../graphql-operations'

export interface BatchSpecStateBadgeProps {
    state: BatchSpecState
}

export const BatchSpecStateBadge: React.FunctionComponent<BatchSpecStateBadgeProps> = ({ state }) => {
    switch (state) {
        case BatchSpecState.PENDING:
            return (
                <Badge variant="secondary" tooltip="Awaiting execution to start">
                    {state}
                </Badge>
            )
        case BatchSpecState.QUEUED:
            return (
                <Badge variant="secondary" tooltip="Waiting for executor">
                    {state}
                </Badge>
            )
        case BatchSpecState.PROCESSING:
            return (
                <Badge variant="secondary" tooltip="Currently executing">
                    {state}
                </Badge>
            )
        case BatchSpecState.CANCELED:
            return (
                <Badge variant="secondary" tooltip="Execution has been canceled">
                    {state}
                </Badge>
            )
        case BatchSpecState.CANCELING:
            return (
                <Badge variant="secondary" tooltip="Canceling execution">
                    {state}
                </Badge>
            )
        case BatchSpecState.FAILED:
            return (
                <Badge variant="danger" tooltip="Execution failed">
                    {state}
                </Badge>
            )
        case BatchSpecState.COMPLETED:
            return (
                <Badge variant="success" tooltip="Execution finished successfully">
                    {state}
                </Badge>
            )
    }
}
