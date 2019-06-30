import { StatusCompletion, StatusResult } from '@sourcegraph/extension-api-classes'
import AlertCircleOutlineIcon from 'mdi-react/AlertCircleOutlineIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import CircleIcon from 'mdi-react/CircleIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import ProgressClockIcon from 'mdi-react/ProgressClockIcon'
import React from 'react'
import * as sourcegraph from 'sourcegraph'

/**
 * Reports whether the status with the given state is completed.
 */
export const statusStateIsCompleted = (
    state: sourcegraph.Status['state']
): state is { completion: sourcegraph.StatusCompletion; result: sourcegraph.StatusResult } =>
    state.completion === StatusCompletion.Completed

type ThemeColor = 'success' | 'failure' | 'muted' | 'info'

export const themeColorForStatus = (status: Pick<sourcegraph.Status, 'state'>): ThemeColor => {
    if (statusStateIsCompleted(status.state)) {
        switch (status.state.result) {
            case StatusResult.Success:
                return 'success'
            case StatusResult.Failure:
                return 'failure'
            case StatusResult.Neutral:
                return 'muted'
            case StatusResult.ActionRequired:
                return 'info'
        }
    }
    return 'muted'
}

/**
 * Returns the icon and theme color class to use for a status.
 */
export const iconForStatus = (
    status: Pick<sourcegraph.Status, 'state'>
): { icon: React.ComponentType<{ className?: string }>; className: string } => {
    const className = `text-${themeColorForStatus(status)}`
    if (statusStateIsCompleted(status.state)) {
        switch (status.state.result) {
            case StatusResult.Success:
                return { icon: CheckIcon, className }
            case StatusResult.Failure:
                return { icon: CloseIcon, className }
            case StatusResult.Neutral:
                return { icon: CircleIcon, className }
            case StatusResult.ActionRequired:
                return { icon: AlertCircleOutlineIcon, className }
        }
    }
    return { icon: ProgressClockIcon, className }
}
