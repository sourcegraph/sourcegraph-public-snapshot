import React from 'react'

import { VisuallyHidden } from '@reach/visually-hidden'

import { Badge, type BadgeVariantType } from '@sourcegraph/wildcard'

import { BatchSpecState } from '../../../../graphql-operations'

export interface BatchSpecStateBadgeProps {
    className?: string
    state: BatchSpecState
}

const getProps = (state: BatchSpecState): [variant: BadgeVariantType, tooltip: string] => {
    switch (state) {
        case BatchSpecState.PENDING: {
            return ['secondary', 'Execution has not been started.']
        }
        case BatchSpecState.QUEUED: {
            return ['secondary', 'Waiting for the next available executor.']
        }
        case BatchSpecState.PROCESSING: {
            return ['secondary', 'The batch spec is actively being executed.']
        }
        case BatchSpecState.CANCELED: {
            return ['secondary', 'Execution of this batch spec has been canceled.']
        }
        case BatchSpecState.CANCELING: {
            return ['secondary', 'Execution of this batch spec is being canceled.']
        }
        case BatchSpecState.FAILED: {
            return [
                'danger',
                "Execution didn't finish successfully in all workspaces. Some changesets may be missing in preview.",
            ]
        }
        case BatchSpecState.COMPLETED: {
            return ['success', 'Execution finished successfully in all workspaces.']
        }
    }
}

export const BatchSpecStateBadge: React.FunctionComponent<React.PropsWithChildren<BatchSpecStateBadgeProps>> = ({
    className,
    state,
}) => {
    const [variant, tooltip] = getProps(state)

    return (
        <>
            <VisuallyHidden role="status">{tooltip}</VisuallyHidden>
            <Badge className={className} variant={variant} tooltip={tooltip} aria-hidden={true}>
                {state}
            </Badge>
        </>
    )
}
