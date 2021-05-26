import classNames from 'classnames'
import { MdiReactIconComponentType } from 'mdi-react'
import WarningIcon from 'mdi-react/AlertCircleOutlineIcon'
import ErrorIcon from 'mdi-react/AlertDecagramOutlineIcon'
import InformationIcon from 'mdi-react/InfoCircleOutlineIcon'
import React from 'react'
import { HoverAlert } from 'sourcegraph'

import { NotificationType } from '../../api/extension/extensionHostApi'
import { renderMarkdown } from '../../util/markdown'
import { useRedesignToggle } from '../../util/useRedesignToggle'
import { GetAlertClassName } from '../HoverOverlay.types'

export interface HoverOverlayAlertsProps {
    hoverAlerts: HoverAlert[]
    iconClassName?: string
    /** Called when an alert is dismissed, with the type of the dismissed alert. */
    onAlertDismissed?: (alertType: string) => void
    getAlertClassName?: GetAlertClassName
}

const hoverAlertIconComponents: Record<Required<HoverAlert>['iconKind'], MdiReactIconComponentType> = {
    info: InformationIcon,
    warning: WarningIcon,
    error: ErrorIcon,
}

function hoverAlertIconComponent(
    iconKind?: Required<HoverAlert>['iconKind'],
    className?: string
): JSX.Element | undefined {
    const PredefinedIcon = iconKind && hoverAlertIconComponents[iconKind]

    return PredefinedIcon && <PredefinedIcon className={classNames('hover-overlay__alert-icon', className)} />
}

const iconKindToNotificationType: Record<Required<HoverAlert>['iconKind'], Parameters<GetAlertClassName>[0]> = {
    info: NotificationType.Info,
    warning: NotificationType.Warning,
    error: NotificationType.Error,
}

export const HoverOverlayAlerts: React.FunctionComponent<HoverOverlayAlertsProps> = props => {
    const { hoverAlerts, iconClassName, onAlertDismissed, getAlertClassName = () => undefined } = props

    const [isRedesignEnabled] = useRedesignToggle()

    const createHandleAlertDismissed = (alertType: string) => (event: React.MouseEvent<HTMLAnchorElement>) => {
        event.preventDefault()

        if (onAlertDismissed) {
            onAlertDismissed(alertType)
        }
    }

    return (
        <div className="hover-overlay__alerts">
            {hoverAlerts.map(({ summary, iconKind, type }, index) => (
                <div
                    key={index}
                    className={classNames(
                        isRedesignEnabled ? 'hover-overlay__alert-redesign' : 'hover-overlay__alert',
                        isRedesignEnabled
                            ? getAlertClassName(iconKind ? iconKindToNotificationType[iconKind] : NotificationType.Info)
                            : getAlertClassName(NotificationType.Info)
                    )}
                >
                    {!isRedesignEnabled && hoverAlertIconComponent(iconKind, iconClassName)}
                    {summary.kind === 'plaintext' ? (
                        <span className="hover-overlay__content">{summary.value}</span>
                    ) : (
                        <span
                            className="hover-overlay__content"
                            dangerouslySetInnerHTML={{ __html: renderMarkdown(summary.value) }}
                        />
                    )}

                    {/* Show dismiss button when an alert has a dismissal type. */}
                    {/* If no type is provided, the alert is not dismissible. */}
                    {type && (
                        <div className="hover-overlay__alert-dismiss">
                            {/* Ideally this should a <button> but we can't guarantee we have the .btn-link class here. */}
                            {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                            <a href="" onClick={createHandleAlertDismissed(type)} role="button">
                                <small>Dismiss</small>
                            </a>
                        </div>
                    )}
                </div>
            ))}
        </div>
    )
}
