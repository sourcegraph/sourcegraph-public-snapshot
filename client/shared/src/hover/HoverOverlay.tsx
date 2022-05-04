import React, { CSSProperties } from 'react'

import classNames from 'classnames'

import { isErrorLike, sanitizeClass } from '@sourcegraph/common'
import { Card } from '@sourcegraph/wildcard'

import { ActionItem, ActionItemComponentProps } from '../actions/ActionItem'
import { NotificationType } from '../api/extension/extensionHostApi'
import { PlatformContextProps } from '../platform/context'
import { TelemetryProps } from '../telemetry/telemetryService'
import { ThemeProps } from '../theme'

import type { HoverContext, HoverOverlayBaseProps, GetAlertClassName, GetAlertVariant } from './HoverOverlay.types'
import { HoverOverlayAlerts, HoverOverlayAlertsProps } from './HoverOverlayAlerts'
import { HoverOverlayContents } from './HoverOverlayContents'
import { HoverOverlayLogo } from './HoverOverlayLogo'
import { useLogTelemetryEvent } from './useLogTelemetryEvent'

import hoverOverlayStyle from './HoverOverlay.module.scss'
import style from './HoverOverlayContents.module.scss'

const LOADING = 'loading' as const

export type { HoverContext }

export interface HoverOverlayClassProps {
    /** An optional class name to apply to the outermost element of the HoverOverlay */
    className?: string

    iconClassName?: string
    badgeClassName?: string

    actionItemClassName?: string
    actionItemPressedClassName?: string

    contentClassName?: string

    /**
     * Allows providing any custom className to style the notifications as desired.
     */
    getAlertClassName?: GetAlertClassName

    /**
     * Allows providing a specific variant style for use in branded Sourcegraph applications.
     */
    getAlertVariant?: GetAlertVariant
}

export interface HoverOverlayProps
    extends HoverOverlayBaseProps,
        ActionItemComponentProps,
        HoverOverlayClassProps,
        TelemetryProps,
        ThemeProps,
        Pick<HoverOverlayAlertsProps, 'onAlertDismissed'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings'> {
    /** A ref callback to get the root overlay element. Use this to calculate the position. */
    hoverRef?: React.Ref<HTMLDivElement>

    /** Show Sourcegraph logo alongside prompt */
    useBrandedLogo?: boolean
}

const getOverlayStyle = (overlayPosition: HoverOverlayProps['overlayPosition']): CSSProperties => {
    if (!overlayPosition) {
        return {
            opacity: 0,
            visibility: 'hidden',
        }
    }

    const topOrBottom = 'top' in overlayPosition ? 'top' : 'bottom'
    const topOrBottomValue = 'top' in overlayPosition ? overlayPosition.top : overlayPosition.bottom

    return {
        opacity: 1,
        visibility: 'visible',
        left: `${overlayPosition.left}px`,
        [topOrBottom]: `${topOrBottomValue}px`,
    }
}

export const HoverOverlay: React.FunctionComponent<React.PropsWithChildren<HoverOverlayProps>> = props => {
    const {
        hoverOrError,
        hoverRef,
        overlayPosition,
        actionsOrError,
        platformContext,
        telemetryService,
        extensionsController,
        location,

        className,
        iconClassName,
        badgeClassName,
        actionItemClassName,
        actionItemPressedClassName,
        contentClassName,

        actionItemStyleProps,

        getAlertClassName,
        getAlertVariant,
        onAlertDismissed,

        useBrandedLogo,
    } = props

    useLogTelemetryEvent(props)

    if (!hoverOrError && (!actionsOrError || isErrorLike(actionsOrError))) {
        return null
    }

    return (
        <Card
            // needed for dynamic styling
            data-testid="hover-overlay"
            // eslint-disable-next-line react/forbid-dom-props
            style={getOverlayStyle(overlayPosition)}
            className={classNames(hoverOverlayStyle.card, hoverOverlayStyle.hoverOverlay, className)}
            ref={hoverRef}
        >
            <div
                data-testid="hover-overlay-contents"
                className={classNames(
                    style.hoverOverlayContents,
                    hoverOrError === LOADING && style.hoverOverlayContentsLoading
                )}
            >
                <HoverOverlayContents
                    hoverOrError={hoverOrError}
                    iconClassName={iconClassName}
                    badgeClassName={badgeClassName}
                    errorAlertClassName={getAlertClassName?.(NotificationType.Error)}
                    errorAlertVariant={getAlertVariant?.(NotificationType.Error)}
                    contentClassName={contentClassName}
                />
            </div>
            {hoverOrError &&
                hoverOrError !== LOADING &&
                !isErrorLike(hoverOrError) &&
                hoverOrError.alerts &&
                hoverOrError.alerts.length > 0 && (
                    <HoverOverlayAlerts
                        hoverAlerts={hoverOrError.alerts}
                        iconClassName={iconClassName}
                        getAlertClassName={getAlertClassName}
                        getAlertVariant={getAlertVariant}
                        onAlertDismissed={onAlertDismissed}
                    />
                )}
            {actionsOrError !== undefined &&
                actionsOrError !== null &&
                actionsOrError !== LOADING &&
                !isErrorLike(actionsOrError) &&
                actionsOrError.length > 0 && (
                    <div className={hoverOverlayStyle.actions}>
                        <div className={hoverOverlayStyle.actionsInner}>
                            {actionsOrError.map((action, index) => (
                                <ActionItem
                                    key={index}
                                    {...action}
                                    className={classNames(
                                        hoverOverlayStyle.action,
                                        actionItemClassName,
                                        `test-tooltip-${sanitizeClass(action.action.title || 'untitled')}`
                                    )}
                                    iconClassName={iconClassName}
                                    pressedClassName={actionItemPressedClassName}
                                    variant="actionItem"
                                    disabledDuringExecution={true}
                                    showLoadingSpinnerDuringExecution={true}
                                    showInlineError={true}
                                    platformContext={platformContext}
                                    telemetryService={telemetryService}
                                    extensionsController={extensionsController}
                                    location={location}
                                    actionItemStyleProps={actionItemStyleProps}
                                />
                            ))}
                        </div>

                        {useBrandedLogo && <HoverOverlayLogo className={hoverOverlayStyle.overlayLogo} />}
                    </div>
                )}
        </Card>
    )
}
