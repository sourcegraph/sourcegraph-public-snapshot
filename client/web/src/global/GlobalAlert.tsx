import React from 'react'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { AlertType } from '@sourcegraph/shared/src/graphql-operations'
import * as GQL from '@sourcegraph/shared/src/schema'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'

import { DismissibleAlert, DismissibleAlertProps } from '../components/DismissibleAlert'

/**
 * A global alert that is shown at the top of the viewport.
 */
export const GlobalAlert: React.FunctionComponent<{
    alert: Pick<GQL.IAlert, 'message' | 'isDismissibleWithKey' | 'type'>
    className: string
}> = ({ alert, className: commonClassName }) => {
    const content = <Markdown dangerousInnerHTML={renderMarkdown(alert.message)} />
    const className = `${commonClassName} alert d-flex`
    if (alert.isDismissibleWithKey) {
        return (
            <DismissibleAlert
                partialStorageKey={`alert.${alert.isDismissibleWithKey}`}
                className={className}
                variant={alertVariantForType(alert.type)}
            >
                {content}
            </DismissibleAlert>
        )
    }
    return <div className={className}>{content}</div>
}

function alertVariantForType(type: AlertType): DismissibleAlertProps['variant'] {
    switch (type) {
        case AlertType.INFO:
            return 'info'
        case AlertType.WARNING:
            return 'warning'
        case AlertType.ERROR:
            return 'danger'
        default:
            return 'warning'
    }
}
