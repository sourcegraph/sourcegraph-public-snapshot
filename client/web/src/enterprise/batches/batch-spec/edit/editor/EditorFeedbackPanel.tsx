import React from 'react'

import classNames from 'classnames'
import { compact } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'

import { ErrorMessage } from '@sourcegraph/branded/src/components/alerts'
import { Icon, Typography } from '@sourcegraph/wildcard'

import { BatchSpecContextErrors } from '../../BatchSpecContext'

import styles from './EditorFeedbackPanel.module.scss'

interface EditorFeedbackPanelProps {
    errors: BatchSpecContextErrors
}

export const EditorFeedbackPanel: React.FunctionComponent<React.PropsWithChildren<EditorFeedbackPanelProps>> = ({
    errors,
}) => {
    const compactedErrors = compact<string | Error>(Object.values(errors))
    if (compactedErrors.length === 0) {
        return null
    }

    const errorHeading = errors.codeValidation ? 'Validation Errors' : 'Errors found'

    return (
        <div className={classNames(styles.panel, 'rounded border bg-1 p-2 w-100 mt-2')}>
            <Typography.H4 className="text-danger text-uppercase">
                <Icon className="text-danger" as={AlertCircleIcon} /> {errorHeading}
            </Typography.H4>
            {compactedErrors.map(error => (
                <ErrorMessage className="text-monospace" error={error} key={String(error)} />
            ))}
        </div>
    )
}
