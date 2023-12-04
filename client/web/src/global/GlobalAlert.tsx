import React from 'react'

import classNames from 'classnames'

import { renderMarkdown } from '@sourcegraph/common'
import { AlertType } from '@sourcegraph/shared/src/graphql-operations'
import { Alert, Markdown } from '@sourcegraph/wildcard'

import { DismissibleAlert, type DismissibleAlertProps } from '../components/DismissibleAlert'
import type { SiteFlagAlertFields } from '../graphql-operations'

/**
 * A global alert that is shown at the top of the viewport.
 */
export const GlobalAlert: React.FunctionComponent<
    React.PropsWithChildren<{
        alert: SiteFlagAlertFields
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
        case AlertType.INFO: {
            return 'info'
        }
        case AlertType.WARNING: {
            return 'warning'
        }
        case AlertType.ERROR: {
            return 'danger'
        }
        default: {
            return 'warning'
        }
    }
}
