import React from 'react'

import classNames from 'classnames'
import { compact } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'

import { ErrorMessage } from '@sourcegraph/branded/src/components/alerts'
import { Icon } from '@sourcegraph/wildcard'

import styles from './EditorFeedbackPanel.module.scss'

interface EditorFeedbackPanelProps {
    errors: {
        codeUpdate: string | Error | undefined
        codeValidation: string | Error | undefined
        preview: string | Error | undefined
        execute: string | Error | undefined
    }
}

export const EditorFeedbackPanel: React.FunctionComponent<EditorFeedbackPanelProps> = ({ errors }) => {
    const compactedErrors = compact(Object.values(errors))
    if (compactedErrors.length === 0) {
        return null
    }

    const errorHeading = errors.codeValidation ? 'Validation Errors' : 'Errors found'

    return (
        <div className={classNames(styles.panel, 'rounded border bg-1 p-2 w-100 mt-2')}>
            <h4 className="text-danger text-uppercase">
                <Icon className="text-danger" as={AlertCircleIcon} /> {errorHeading}
            </h4>
            {compactedErrors.map(error => (
                <ErrorMessage className="text-monospace" error={error} key={String(error)} />
            ))}
        </div>
    )
}
