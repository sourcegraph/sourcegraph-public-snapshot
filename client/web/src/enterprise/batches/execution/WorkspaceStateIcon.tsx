import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import ContentSaveIcon from 'mdi-react/ContentSaveIcon'
import LinkVariantRemoveIcon from 'mdi-react/LinkVariantRemoveIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'
import React from 'react'

import { LoadingSpinner } from '@sourcegraph/wildcard'

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
                <TimerSandIcon
                    className={classNames('icon-inline text-muted', className)}
                    data-tooltip="Waiting to be processed"
                />
            )
        case BatchSpecWorkspaceState.PROCESSING:
            return (
                <LoadingSpinner
                    className={classNames('icon-inline text-muted', className)}
                    data-tooltip="Currently executing"
                />
            )
        case BatchSpecWorkspaceState.SKIPPED:
            return (
                <LinkVariantRemoveIcon
                    className={classNames('icon-inline text-muted', className)}
                    data-tooltip="The workspace has been skipped"
                />
            )
        case BatchSpecWorkspaceState.CANCELED:
            return (
                <AlertCircleIcon
                    className={classNames('icon-inline text-muted', className)}
                    data-tooltip="The execution has been canceled"
                />
            )
        case BatchSpecWorkspaceState.CANCELING:
            return (
                <AlertCircleIcon
                    className={classNames('icon-inline text-muted', className)}
                    data-tooltip="The execution is being canceled"
                />
            )
        case BatchSpecWorkspaceState.FAILED:
            return (
                <AlertCircleIcon
                    className={classNames('icon-inline text-danger', className)}
                    data-tooltip="The execution has failed"
                />
            )
        case BatchSpecWorkspaceState.COMPLETED:
            if (cachedResultFound) {
                return (
                    <ContentSaveIcon
                        className={classNames('icon-inline text-success', className)}
                        data-tooltip="Cached result found"
                    />
                )
            }
            return (
                <CheckCircleIcon
                    className={classNames('icon-inline text-success', className)}
                    data-tooltip="Successfully executed"
                />
            )
    }
}
