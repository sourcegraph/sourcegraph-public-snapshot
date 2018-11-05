import marked from 'marked'
import ErrorIcon from 'mdi-react/ErrorIcon'
import InformationIcon from 'mdi-react/InformationIcon'
import WarningIcon from 'mdi-react/WarningIcon'
import React from 'react'
import * as GQL from '../backend/graphqlschema'
import { DismissibleAlert } from '../components/DismissibleAlert'
import { Markdown } from '../components/Markdown'

/**
 * A global alert that is shown at the top of the viewport.
 */
export const GlobalAlert: React.SFC<{ alert: GQL.IAlert; className: string }> = ({
    alert,
    className: commonClassName,
}) => {
    const Icon = alertIconForType(alert.type)
    const content = (
        <>
            <Icon className="icon-inline mr-2 flex-shrink-0" />
            <Markdown dangerousInnerHTML={marked(alert.message, { gfm: true, breaks: true, sanitize: true })} />
        </>
    )
    const className = `${commonClassName} alert alert-${alertClassForType(alert.type)} d-flex`
    if (alert.isDismissibleWithKey) {
        return (
            <DismissibleAlert partialStorageKey={`alert.${alert.isDismissibleWithKey}`} className={className}>
                {content}
            </DismissibleAlert>
        )
    }
    return <div className={className}>{content}</div>
}

function alertIconForType(type: GQL.AlertType): React.ComponentType<{ className?: string }> {
    switch (type) {
        case GQL.AlertType.INFO:
            return InformationIcon
        case GQL.AlertType.WARNING:
            return WarningIcon
        case GQL.AlertType.ERROR:
            return ErrorIcon
        default:
            return WarningIcon
    }
}

function alertClassForType(type: GQL.AlertType): string {
    switch (type) {
        case GQL.AlertType.INFO:
            return 'info'
        case GQL.AlertType.WARNING:
            return 'warning'
        case GQL.AlertType.ERROR:
            return 'danger'
        default:
            return 'warning'
    }
}
