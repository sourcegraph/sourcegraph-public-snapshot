import { NotificationType } from '@sourcegraph/extension-api-classes'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import AlertIcon from 'mdi-react/AlertIcon'
import HelpCircleOutlineIcon from 'mdi-react/HelpCircleOutlineIcon'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'
import React from 'react'
import { NotificationType as NotificationTypeType } from 'sourcegraph'

const DEFAULT_TYPE: NotificationTypeType = NotificationType.Info

interface SeverityInfo {
    icon: React.ComponentType<{ className?: string }>
    ariaLabel: string
    className: string
}

const INFO: Record<NotificationTypeType, SeverityInfo> = {
    [NotificationType.Error]: { icon: AlertCircleIcon, ariaLabel: 'Error', className: 'text-danger' },
    [NotificationType.Warning]: { icon: AlertIcon, ariaLabel: 'Warning', className: 'text-warning' },
    [NotificationType.Info]: { icon: InformationOutlineIcon, ariaLabel: 'Info', className: 'text-info' },
    [NotificationType.Success]: { icon: HelpCircleOutlineIcon, ariaLabel: 'Success', className: 'text-success' },
    [NotificationType.Log]: { icon: HelpCircleOutlineIcon, ariaLabel: 'Log', className: 'text-muted' },
}

/**
 * An icon representing the type of a notification.
 */
export const NotificationTypeIcon: React.FunctionComponent<{
    type: NotificationTypeType
    className?: string
}> = ({ type, className = '' }) => {
    const { icon: Icon, ariaLabel, className: typeClassName } = INFO[type] || INFO[DEFAULT_TYPE]
    return <Icon className={`${typeClassName} ${className}`} aria-label={ariaLabel} />
}
