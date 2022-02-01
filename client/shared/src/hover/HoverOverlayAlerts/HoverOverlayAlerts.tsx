import classNames from 'classnames'
import React from 'react'
import { HoverAlert } from 'sourcegraph'

import { Alert } from '@sourcegraph/wildcard'

import { NotificationType } from '../../api/extension/extensionHostApi'
import { renderMarkdown } from '../../util/markdown'
import hoverOverlayStyle from '../HoverOverlay.module.scss'
import { GetAlertClassName, GetAlertVariant } from '../HoverOverlay.types'
import contentStyles from '../HoverOverlayContents/HoverOverlayContent/HoverOverlayContent.module.scss'

import styles from './HoverOverlayAlerts.module.scss'

export interface HoverOverlayAlertsProps {
    hoverAlerts: HoverAlert[]
    iconClassName?: string
    /** Called when an alert is dismissed, with the type of the dismissed alert. */
    onAlertDismissed?: (alertType: string) => void
    getAlertClassName?: GetAlertClassName
    getAlertVariant?: GetAlertVariant
    className?: string
}

const iconKindToNotificationType: Record<Required<HoverAlert>['iconKind'], Parameters<GetAlertClassName>[0]> = {
    info: NotificationType.Info,
    warning: NotificationType.Warning,
    error: NotificationType.Error,
}

export const HoverOverlayAlerts: React.FunctionComponent<HoverOverlayAlertsProps> = props => {
    const { hoverAlerts, onAlertDismissed, getAlertClassName, getAlertVariant } = props

    const createAlertDismissedHandler = (alertType: string) => (event: React.MouseEvent<HTMLAnchorElement>) => {
        event.preventDefault()

        if (onAlertDismissed) {
            onAlertDismissed(alertType)
        }
    }

    return (
        <div className={classNames(styles.hoverOverlayAlerts, props.className)}>
            {hoverAlerts.map(({ summary, iconKind, type }, index) => (
                <Alert
                    key={index}
                    variant={getAlertVariant?.(iconKind ? iconKindToNotificationType[iconKind] : NotificationType.Info)}
                    className={classNames(
                        hoverOverlayStyle.alert,
                        getAlertClassName?.(iconKind ? iconKindToNotificationType[iconKind] : NotificationType.Info)
                    )}
                >
                    {summary.kind === 'plaintext' ? (
                        <span
                            data-testid="hover-overlay-content"
                            className={classNames(
                                contentStyles.hoverOverlayContent,
                                hoverOverlayStyle.hoverOverlayContent
                            )}
                        >
                            {summary.value}
                        </span>
                    ) : (
                        <span
                            data-testid="hover-overlay-content"
                            className={classNames(
                                contentStyles.hoverOverlayContent,
                                hoverOverlayStyle.hoverOverlayContent
                            )}
                            dangerouslySetInnerHTML={{ __html: renderMarkdown(summary.value) }}
                        />
                    )}

                    {/* Show dismiss button when an alert has a dismissal type. */}
                    {/* If no type is provided, the alert is not dismissible. */}
                    {type && (
                        <div className={classNames(hoverOverlayStyle.alertDismiss)}>
                            {/* Ideally this should a <button> but we can't guarantee we have the .btn-link class here. */}
                            {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                            <a href="" onClick={createAlertDismissedHandler(type)} role="button">
                                <small>Dismiss</small>
                            </a>
                        </div>
                    )}
                </Alert>
            ))}
        </div>
    )
}
