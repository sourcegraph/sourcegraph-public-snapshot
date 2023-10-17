import React from 'react'

import { mdiAlertCircle } from '@mdi/js'
import classNames from 'classnames'
import { compact } from 'lodash'

import { Icon, H4, ErrorMessage } from '@sourcegraph/wildcard'

import type { BatchSpecContextErrors } from '../../BatchSpecContext'

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
        <div
            className={classNames(styles.panel, 'rounded border bg-1 p-2 w-100 mt-2')}
            role="region"
            aria-label="editor feedback panel"
        >
            <H4 className="text-danger text-uppercase">
                <Icon aria-hidden={true} className="text-danger" svgPath={mdiAlertCircle} /> {errorHeading}
            </H4>
            {compactedErrors.map(error => (
                <ErrorMessage className="text-monospace" error={error} key={String(error)} />
            ))}
        </div>
    )
}
