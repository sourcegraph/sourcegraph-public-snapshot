import React from 'react'

import { mdiTimerSand, mdiLinkVariantRemove, mdiCancel, mdiAlertCircle, mdiContentSave, mdiCheckBold } from '@mdi/js'
import classNames from 'classnames'

import { LoadingSpinner, Icon } from '@sourcegraph/wildcard'

import { BatchSpecWorkspaceState } from '../../../../../graphql-operations'

export interface WorkspaceStateIconProps {
    state: BatchSpecWorkspaceState
    cachedResultFound: boolean
    className?: string
}

export const WorkspaceStateIcon: React.FunctionComponent<React.PropsWithChildren<WorkspaceStateIconProps>> = ({
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
                    data-tooltip="This workspace is queued for execution."
                    aria-label="This workspace is queued for execution."
                    svgPath={mdiTimerSand}
                />
            )
        case BatchSpecWorkspaceState.PROCESSING:
            return (
                <Icon
                    className={classNames('text-muted', className)}
                    data-tooltip="This workspace is currently executing."
                    aria-label="This workspace is currently executing."
                    as={LoadingSpinner}
                />
            )
        case BatchSpecWorkspaceState.SKIPPED:
            return (
                <Icon
                    className={classNames('text-muted', className)}
                    data-tooltip="This workspace was skipped."
                    aria-label="This workspace was skipped."
                    svgPath={mdiLinkVariantRemove}
                />
            )
        case BatchSpecWorkspaceState.CANCELED:
            return (
                <Icon
                    className={classNames('text-muted', className)}
                    data-tooltip="The execution for this workspace was canceled."
                    aria-label="The execution for this workspace was canceled."
                    svgPath={mdiCancel}
                />
            )
        case BatchSpecWorkspaceState.CANCELING:
            return (
                <Icon
                    className={classNames('text-muted', className)}
                    data-tooltip="The execution for this workspace is being canceled."
                    aria-label="The execution for this workspace is being canceled."
                    svgPath={mdiCancel}
                />
            )
        case BatchSpecWorkspaceState.FAILED:
            return (
                <Icon
                    className="text-danger"
                    data-tooltip="The execution for this workspace failed."
                    aria-label="The execution for this workspace failed."
                    svgPath={mdiAlertCircle}
                />
            )
        case BatchSpecWorkspaceState.COMPLETED:
            if (cachedResultFound) {
                return (
                    <Icon
                        className="text-success"
                        data-tooltip="Cached result found for this workspace."
                        aria-label="Cached result found for this workspace."
                        svgPath={mdiContentSave}
                    />
                )
            }
            return (
                <Icon
                    className="text-success"
                    data-tooltip="Execution for this workspace succeeded."
                    aria-label="Execution for this workspace succeeded."
                    svgPath={mdiCheckBold}
                />
            )
    }
}
