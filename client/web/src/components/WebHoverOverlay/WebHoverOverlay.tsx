import React, { useCallback, useEffect } from 'react'

import { HoverOverlay, HoverOverlayProps } from '@sourcegraph/shared/src/hover/HoverOverlay'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'

import { HoverThresholdProps } from '../../repo/RepoContainer'

const iconKindToAlertKind = {
    info: 'secondary',
    error: 'danger',
    warning: 'warning',
}

const getAlertClassName: HoverOverlayProps['getAlertClassName'] = iconKind =>
    `alert alert-${iconKindToAlertKind[iconKind]}`

export const WebHoverOverlay: React.FunctionComponent<HoverOverlayProps & HoverThresholdProps> = props => {
    const [dismissedAlerts, setDismissedAlerts] = useLocalStorage<string[]>('WebHoverOverlay.dismissedAlerts', [])
    const onAlertDismissed = useCallback(
        (alertType: string) => {
            if (!dismissedAlerts.includes(alertType)) {
                setDismissedAlerts([...dismissedAlerts, alertType])
            }
        },
        [dismissedAlerts, setDismissedAlerts]
    )

    let propsToUse = props
    if (props.hoverOrError && props.hoverOrError !== 'loading' && !isErrorLike(props.hoverOrError)) {
        const filteredAlerts = (props.hoverOrError?.alerts || []).filter(
            alert => !alert.type || !dismissedAlerts.includes(alert.type)
        )
        propsToUse = { ...props, hoverOrError: { ...props.hoverOrError, alerts: filteredAlerts } }
    }

    const { hoverOrError } = propsToUse
    const { onHoverShown, hoveredToken } = props

    /** Whether the hover has actual content (that provides value to the user) */
    const hoverHasValue = hoverOrError !== 'loading' && !isErrorLike(hoverOrError) && !!hoverOrError?.contents?.length

    useEffect(() => {
        if (hoverHasValue) {
            onHoverShown?.()
        }
    }, [hoveredToken?.filePath, hoveredToken?.line, hoveredToken?.character, onHoverShown, hoverHasValue])

    return (
        <HoverOverlay
            {...propsToUse}
            className="card"
            iconClassName="icon-inline"
            closeButtonClassName="btn btn-icon"
            actionItemClassName="btn btn-secondary"
            onAlertDismissed={onAlertDismissed}
            getAlertClassName={getAlertClassName}
        />
    )
}

WebHoverOverlay.displayName = 'WebHoverOverlay'
