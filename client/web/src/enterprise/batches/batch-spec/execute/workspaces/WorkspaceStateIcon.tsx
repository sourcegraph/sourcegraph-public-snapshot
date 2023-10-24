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
        case BatchSpecWorkspaceState.PENDING: {
            return null
        }
        case BatchSpecWorkspaceState.QUEUED: {
            return (
                <Tooltip content="This workspace is queued for execution.">
                    <Icon
                        aria-label="This workspace is queued for execution."
                        className={classNames('text-muted', className)}
                        svgPath={mdiTimerSand}
                    />
                </Tooltip>
            )
        }
        case BatchSpecWorkspaceState.PROCESSING: {
            return (
                <Tooltip content="This workspace is currently executing.">
                    <span>
                        <Icon
                            aria-label="This workspace is currently executing."
                            className={className}
                            as={LoadingSpinner}
                            aria-live="off"
                        />
                    </span>
                </Tooltip>
            )
        }
        case BatchSpecWorkspaceState.SKIPPED: {
            return (
                <Tooltip content="This workspace was skipped.">
                    <Icon
                        aria-label="This workspace was skipped."
                        className={classNames('text-muted', className)}
                        svgPath={mdiLinkVariantRemove}
                    />
                </Tooltip>
            )
        }
        case BatchSpecWorkspaceState.CANCELED: {
            return (
                <Tooltip content="The execution for this workspace was canceled.">
                    <Icon
                        aria-label="The execution for this workspace was canceled."
                        className={classNames('text-muted', className)}
                        svgPath={mdiCancel}
                    />
                </Tooltip>
            )
        }
        case BatchSpecWorkspaceState.CANCELING: {
            return (
                <Tooltip content="The execution for this workspace is being canceled.">
                    <Icon
                        aria-label="The execution for this workspace is being canceled."
                        className={classNames('text-muted', className)}
                        svgPath={mdiCancel}
                    />
                </Tooltip>
            )
        }
        case BatchSpecWorkspaceState.FAILED: {
            return (
                <Tooltip content="The execution for this workspace failed.">
                    <Icon
                        aria-label="The execution for this workspace failed."
                        className="text-danger"
                        svgPath={mdiAlertCircle}
                    />
                </Tooltip>
            )
        }
        case BatchSpecWorkspaceState.COMPLETED: {
            if (cachedResultFound) {
                return (
                    <Tooltip content="Cached result found for this workspace.">
                        <Icon
                            aria-label="Cached result found for this workspace."
                            className="text-success"
                            svgPath={mdiContentSave}
                        />
                    </Tooltip>
                )
            }
            return (
                <Tooltip content="Execution for this workspace succeeded.">
                    <Icon
                        aria-label="Execution for this workspace succeeded."
                        className="text-success"
                        svgPath={mdiCheckBold}
                    />
                </Tooltip>
            )
        }
    }
}
