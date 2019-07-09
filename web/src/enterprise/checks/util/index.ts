import { CheckCompletion, CheckResult } from '@sourcegraph/extension-api-classes'
import AlertCircleOutlineIcon from 'mdi-react/AlertCircleOutlineIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import CircleIcon from 'mdi-react/CircleIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import ProgressClockIcon from 'mdi-react/ProgressClockIcon'
import React from 'react'
import * as sourcegraph from 'sourcegraph'

/**
 * Reports whether the check with the given state is completed.
 */
export const checkStateIsCompleted = (
    state: sourcegraph.CheckInformation['state']
): state is { completion: sourcegraph.CheckCompletion; result: sourcegraph.CheckResult } =>
    state.completion === CheckCompletion.Completed

type ThemeColor = 'success' | 'danger' | 'muted' | 'info'

export const themeColorForCheck = (status: Pick<sourcegraph.CheckInformation, 'state'>): ThemeColor => {
    if (checkStateIsCompleted(status.state)) {
        switch (status.state.result) {
            case CheckResult.Success:
                return 'success'
            case CheckResult.Failure:
                return 'danger'
            case CheckResult.Neutral:
                return 'muted'
            case CheckResult.ActionRequired:
                return 'info'
        }
    }
    return 'muted'
}

/**
 * Returns the icon and theme color class to use for a check.
 */
export const iconForCheck = (
    status: Pick<sourcegraph.CheckInformation, 'state'>
): { icon: React.ComponentType<{ className?: string }>; className: string } => {
    const className = `text-${themeColorForCheck(status)}`
    if (checkStateIsCompleted(status.state)) {
        switch (status.state.result) {
            case CheckResult.Success:
                return { icon: CheckIcon, className }
            case CheckResult.Failure:
                return { icon: CloseIcon, className }
            case CheckResult.Neutral:
                return { icon: CircleIcon, className }
            case CheckResult.ActionRequired:
                return { icon: AlertCircleOutlineIcon, className }
        }
    }
    return { icon: ProgressClockIcon, className }
}
