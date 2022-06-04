import React from 'react'

import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CancelIcon from 'mdi-react/CancelIcon'
import CheckBoldIcon from 'mdi-react/CheckBoldIcon'
import ContentSaveIcon from 'mdi-react/ContentSaveIcon'
import LinkVariantRemoveIcon from 'mdi-react/LinkVariantRemoveIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'

import { LoadingSpinner, Icon } from '@sourcegraph/wildcard'

import { BatchSpecWorkspaceState } from '../../../../../graphql-operations'

export interface WorkspaceStateIconProps {
    state: BatchSpecWorkspaceState
    cachedResultFound: boolean
}

export const WorkspaceStateIcon: React.FunctionComponent<React.PropsWithChildren<WorkspaceStateIconProps>> = ({
    state,
    cachedResultFound,
}) => {
    switch (state) {
        case BatchSpecWorkspaceState.PENDING:
            return null
        case BatchSpecWorkspaceState.QUEUED:
            return (
                <Icon
                    role="img"
                    className="text-muted"
                    data-tooltip="This workspace is queued for execution."
                    aria-label="This workspace is queued for execution."
                    as={TimerSandIcon}
                />
            )
        case BatchSpecWorkspaceState.PROCESSING:
            return (
                <Icon
                    role="img"
                    className="text-muted"
                    data-tooltip="This workspace is currently executing."
                    aria-label="This workspace is currently executing."
                    as={LoadingSpinner}
                />
            )
        case BatchSpecWorkspaceState.SKIPPED:
            return (
                <Icon
                    role="img"
                    className="text-muted"
                    data-tooltip="This workspace was skipped."
                    aria-label="This workspace was skipped."
                    as={LinkVariantRemoveIcon}
                />
            )
        case BatchSpecWorkspaceState.CANCELED:
            return (
                <Icon
                    role="img"
                    className="text-muted"
                    data-tooltip="The execution for this workspace was canceled."
                    aria-label="The execution for this workspace was canceled."
                    as={CancelIcon}
                />
            )
        case BatchSpecWorkspaceState.CANCELING:
            return (
                <Icon
                    role="img"
                    className="text-muted"
                    data-tooltip="The execution for this workspace is being canceled."
                    aria-label="The execution for this workspace is being canceled."
                    as={CancelIcon}
                />
            )
        case BatchSpecWorkspaceState.FAILED:
            return (
                <Icon
                    role="img"
                    className="text-danger"
                    data-tooltip="The execution for this workspace failed."
                    aria-label="The execution for this workspace failed."
                    as={AlertCircleIcon}
                />
            )
        case BatchSpecWorkspaceState.COMPLETED:
            if (cachedResultFound) {
                return (
                    <Icon
                        role="img"
                        className="text-success"
                        data-tooltip="Cached result found for this workspace."
                        aria-label="Cached result found for this workspace."
                        as={ContentSaveIcon}
                    />
                )
            }
            return (
                <Icon
                    role="img"
                    className="text-success"
                    data-tooltip="Execution for this workspace succeeded."
                    aria-label="Execution for this workspace succeeded."
                    as={CheckBoldIcon}
                />
            )
    }
}
