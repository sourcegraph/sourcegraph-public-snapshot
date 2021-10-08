import classNames from 'classnames'
import React from 'react'

import {
    LoadingSpinner as ReactLoadingSpinner,
    LoadingSpinnerProps as ReactLoadingSpinnerProps,
} from '@sourcegraph/react-loading-spinner'

interface LoadingSpinnerProps extends ReactLoadingSpinnerProps {
    /** Determine whether the spinner should be besides any other element or not */
    inline?: boolean
}

/**
 * A simple wrapper around the generic Sourcegraph React loading spinner
 *
 * Supports additional custom styling relevant to this codebase.
 */
export const LoadingSpinner: React.FunctionComponent<LoadingSpinnerProps> = ({ className, inline }) => (
    <ReactLoadingSpinner className={classNames(inline && 'icon-inline', className)} />
)
