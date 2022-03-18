import React from 'react'

import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import ContentSaveIcon from 'mdi-react/ContentSaveIcon'
import LinkVariantRemoveIcon from 'mdi-react/LinkVariantRemoveIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'

import { LoadingSpinner, Icon } from '@sourcegraph/wildcard'

import { BatchSpecWorkspaceState } from '../../../graphql-operations'

export interface WorkspaceStateIconProps {
    state: BatchSpecWorkspaceState
    cachedResultFound: boolean

    className?: string
}

export const WorkspaceStateIcon: React.FunctionComponent<WorkspaceStateIconProps> = ({
    state,
    cachedResultFound,
    className,
}) => {
    switch (state) {
        case BatchSpecWorkspaceState.PENDING:
            return null
        case BatchSpecWorkspaceState.QUEUED:
            return (
                <Icon
                    className={classNames('text-muted', className)}
                    data-tooltip="Waiting to be processed"
                    as={TimerSandIcon}
                />
            )
        case BatchSpecWorkspaceState.PROCESSING:
            return (
                <Icon
                    className={classNames('text-muted', className)}
                    data-tooltip="Currently executing"
                    as={LoadingSpinner}
                />
            )
        case BatchSpecWorkspaceState.SKIPPED:
            return (
                <Icon
                    className={classNames('text-muted', className)}
                    data-tooltip="The workspace has been skipped"
                    as={LinkVariantRemoveIcon}
                />
            )
        case BatchSpecWorkspaceState.CANCELED:
            return (
                <Icon
                    className={classNames('text-muted', className)}
                    data-tooltip="The execution has been canceled"
                    as={AlertCircleIcon}
                />
            )
        case BatchSpecWorkspaceState.CANCELING:
            return (
                <Icon
                    className={classNames('text-muted', className)}
                    data-tooltip="The execution is being canceled"
                    as={AlertCircleIcon}
                />
            )
        case BatchSpecWorkspaceState.FAILED:
            return (
                <Icon
                    className={classNames('text-danger', className)}
                    data-tooltip="The execution has failed"
                    as={AlertCircleIcon}
                />
            )
        case BatchSpecWorkspaceState.COMPLETED:
            if (cachedResultFound) {
                return (
                    <Icon
                        className={classNames('text-success', className)}
                        data-tooltip="Cached result found"
                        as={ContentSaveIcon}
                    />
                )
            }
            return (
                <Icon
                    className={classNames('text-success', className)}
                    data-tooltip="Successfully executed"
                    as={CheckCircleIcon}
                />
            )
    }
}
