import React from 'react'

import classNames from 'classnames'

import { renderMarkdown } from '@sourcegraph/common'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { AlertType } from '@sourcegraph/shared/src/graphql-operations'
import * as GQL from '@sourcegraph/shared/src/schema'
import { Alert } from '@sourcegraph/wildcard'

import { DismissibleAlert, DismissibleAlertProps } from '../components/DismissibleAlert'

/**
 * A global alert that is shown at the top of the viewport.
 */
export const GlobalAlert: React.FunctionComponent<
    React.PropsWithChildren<{
        alert: Pick<GQL.IAlert, 'message' | 'isDismissibleWithKey' | 'type'>
        className?: string
    }>
> = ({ alert, className: commonClassName }) => {
    const content = <Markdown dangerousInnerHTML={renderMarkdown(alert.message)} />
    const className = classNames(commonClassName, 'd-flex')

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
    return (
        <Alert className={className} variant={alertVariantForType(alert.type)}>
            {content}
        </Alert>
    )
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
