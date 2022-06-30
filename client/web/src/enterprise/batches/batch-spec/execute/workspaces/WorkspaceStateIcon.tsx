import React from 'react'

import { mdiTimerSand, mdiLinkVariantRemove, mdiCancel, mdiAlertCircle, mdiContentSave, mdiCheckBold } from '@mdi/js'
import classNames from 'classnames'

import { LoadingSpinner, Icon, Tooltip } from '@sourcegraph/wildcard'

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
                <Tooltip content="This workspace is queued for execution.">
                    <Icon
                        className={classNames('text-muted', className)}
                        aria-label="This workspace is queued for execution."
                        svgPath={mdiTimerSand}
                    />
                </Tooltip>
            )
        case BatchSpecWorkspaceState.PROCESSING:
            return (
                <Tooltip content="This workspace is currently executing.">
                    <Icon
                        className={classNames('text-muted', className)}
                        aria-label="This workspace is currently executing."
                        as={LoadingSpinner}
                    />
                </Tooltip>
            )
        case BatchSpecWorkspaceState.SKIPPED:
            return (
                <Tooltip content="This workspace was skipped.">
                    <Icon
                        className={classNames('text-muted', className)}
                        aria-label="This workspace was skipped."
                        svgPath={mdiLinkVariantRemove}
                    />
                </Tooltip>
            )
        case BatchSpecWorkspaceState.CANCELED:
            return (
                <Tooltip content="The execution for this workspace was canceled.">
                    <Icon
                        className={classNames('text-muted', className)}
                        aria-label="The execution for this workspace was canceled."
                        svgPath={mdiCancel}
                    />
                </Tooltip>
            )
        case BatchSpecWorkspaceState.CANCELING:
            return (
                <Tooltip content="The execution for this workspace is being canceled.">
                    <Icon
                        className={classNames('text-muted', className)}
                        aria-label="The execution for this workspace is being canceled."
                        svgPath={mdiCancel}
                    />
                </Tooltip>
            )
        case BatchSpecWorkspaceState.FAILED:
            return (
                <Tooltip content="The execution for this workspace failed.">
                    <Icon
                        className="text-danger"
                        aria-label="The execution for this workspace failed."
                        svgPath={mdiAlertCircle}
                    />
                </Tooltip>
            )
        case BatchSpecWorkspaceState.COMPLETED:
            if (cachedResultFound) {
                return (
                    <Tooltip content="Cached result found for this workspace.">
                        <Icon
                            className="text-success"
                            aria-label="Cached result found for this workspace."
                            svgPath={mdiContentSave}
                        />
                    </Tooltip>
                )
            }
            return (
                <Tooltip content="Execution for this workspace succeeded.">
                    <Icon
                        className="text-success"
                        aria-label="Execution for this workspace succeeded."
                        svgPath={mdiCheckBold}
                    />
                </Tooltip>
            )
    }
}
