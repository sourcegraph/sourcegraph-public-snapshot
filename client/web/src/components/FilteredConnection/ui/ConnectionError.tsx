import classNames from 'classnames'
import React from 'react'

import { ErrorMessage } from '../../alerts'

import styles from './ConnectionError.module.scss'

interface ConnectionErrorProps {
    errors: string[]
    compact?: boolean
}

/**
 * Renders FilteredConnection styled errors
 */
export const ConnectionError: React.FunctionComponent<ConnectionErrorProps> = ({ errors, compact }) => (
    <div className={classNames('alert alert-danger', compact && styles.compact)}>
        {errors.map((error, index) => (
            <React.Fragment key={index}>
                <ErrorMessage error={error} />
            </React.Fragment>
        ))}
    </div>
)
