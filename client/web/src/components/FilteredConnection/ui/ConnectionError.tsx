import React from 'react'

import classNames from 'classnames'

import { ErrorMessage } from '@sourcegraph/branded/src/components/alerts'
import { Alert } from '@sourcegraph/wildcard'

import styles from './ConnectionError.module.scss'

interface ConnectionErrorProps {
    errors: string[]
    compact?: boolean
}

/**
 * Renders FilteredConnection styled errors
 */
export const ConnectionError: React.FunctionComponent<React.PropsWithChildren<ConnectionErrorProps>> = ({
    errors,
    compact,
}) => (
    <Alert className={classNames(compact && styles.compact)} variant="danger">
        {errors.map((error, index) => (
            <React.Fragment key={index}>
                <ErrorMessage error={error} />
            </React.Fragment>
        ))}
    </Alert>
)
