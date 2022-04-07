import React from 'react'

import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'

import { ErrorMessage } from '@sourcegraph/branded/src/components/alerts'
import { Icon } from '@sourcegraph/wildcard'

import styles from './EditorFeedbackPanel.module.scss'

export const EditorFeedbackPanel: React.FunctionComponent<{ errors: (string | Error)[] }> = ({ errors }) => {
    if (errors.length === 0) {
        return null
    }

    return (
        <div className={classNames(styles.panel, 'rounded border bg-1 p-2 w-100 mt-2')}>
            <h4 className="text-danger text-uppercase">
                <Icon className="text-danger" as={AlertCircleIcon} /> Validation Errors
            </h4>
            {errors.map(error => (
                <ErrorMessage className="text-monospace" error={error} key={String(error)} />
            ))}
        </div>
    )
}
