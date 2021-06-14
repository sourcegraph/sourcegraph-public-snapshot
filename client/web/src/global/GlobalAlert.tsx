import React from 'react'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { AlertType } from '@sourcegraph/shared/src/graphql-operations'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'

import { DismissibleAlert } from '../components/DismissibleAlert'

/**
 * A global alert that is shown at the top of the viewport.
 */
export const GlobalAlert: React.FunctionComponent<{ alert: GQL.IAlert; className: string }> = ({
    alert,
    className: commonClassName,
}) => {
    const content = <Markdown dangerousInnerHTML={renderMarkdown(alert.message)} />
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

function alertClassForType(type: AlertType): string {
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
