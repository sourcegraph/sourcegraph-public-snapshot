import classNames from 'classnames'
import React, { CSSProperties } from 'react'

import { ActionItem, ActionItemComponentProps } from '../actions/ActionItem'
import { NotificationType } from '../api/extension/extensionHostApi'
import { PlatformContextProps } from '../platform/context'
import { TelemetryProps } from '../telemetry/telemetryService'
import { ThemeProps } from '../theme'
import { isErrorLike } from '../util/errors'
import { sanitizeClass } from '../util/strings'

import hoverOverlayStyle from './HoverOverlay.module.scss'
import type { HoverContext, HoverOverlayBaseProps, GetAlertClassName } from './HoverOverlay.types'
import { HoverOverlayAlerts, HoverOverlayAlertsProps } from './HoverOverlayAlerts'
import { HoverOverlayContents } from './HoverOverlayContents'
import style from './HoverOverlayContents.module.scss'
import { HoverOverlayLogo } from './HoverOverlayLogo'
import { useLogTelemetryEvent } from './useLogTelemetryEvent'

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

    getAlertClassName?: GetAlertClassName
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

    /** Show Sourcegraph branded badges */
    useBrandedBadge?: boolean
}

const getOverlayStyle = (overlayPosition: HoverOverlayProps['overlayPosition']): CSSProperties =>
    overlayPosition
        ? {
              opacity: 1,
              visibility: 'visible',
              left: `${overlayPosition.left}px`,
              top: `${overlayPosition.top}px`,
          }
        : {
              opacity: 0,
              visibility: 'hidden',
          }

export const HoverOverlay: React.FunctionComponent<HoverOverlayProps> = props => {
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

        getAlertClassName,
        onAlertDismissed,

        useBrandedLogo,
        useBrandedBadge,
    } = props

    useLogTelemetryEvent(props)

    if (!hoverOrError && (!actionsOrError || isErrorLike(actionsOrError))) {
        return null
    }

    return (
        <div
            // needed for dynamic styling
            data-testid="hover-overlay"
            // eslint-disable-next-line react/forbid-dom-props
            style={getOverlayStyle(overlayPosition)}
            className={classNames(hoverOverlayStyle.hoverOverlay, className)}
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
                    useBrandedBadge={useBrandedBadge}
                    badgeClassName={badgeClassName}
                    errorAlertClassName={getAlertClassName?.(NotificationType.Error)}
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
                                />
                            ))}
                        </div>

                        {useBrandedLogo && <HoverOverlayLogo className={hoverOverlayStyle.overlayLogo} />}
                    </div>
                )}
        </div>
    )
}
