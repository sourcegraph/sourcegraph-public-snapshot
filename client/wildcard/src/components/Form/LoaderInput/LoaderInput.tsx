import React from 'react'

import classNames from 'classnames'

import { LoadingSpinner } from '../../LoadingSpinner/LoadingSpinner'

import styles from './LoaderInput.module.scss'

/** Takes loading prop, input component as child */
interface LoaderInputProps {
    loading: boolean
    children: React.ReactNode
    className?: string
}

export const LoaderInput: React.FunctionComponent<React.PropsWithChildren<LoaderInputProps>> = ({
    loading,
    children,
    className,
}) => (
    <div className={classNames(styles.container, className)}>
        {children}
        {loading && <LoadingSpinner inline={false} className={styles.spinner} />}
    </div>
)
