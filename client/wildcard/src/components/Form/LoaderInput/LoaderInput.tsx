import { type HTMLAttributes, forwardRef } from 'react'

import classNames from 'classnames'

import { LoadingSpinner } from '../../LoadingSpinner'

import styles from './LoaderInput.module.scss'

interface LoaderInputProps extends HTMLAttributes<HTMLDivElement> {
    loading: boolean
    children: React.ReactNode
}

export const LoaderInput = forwardRef<HTMLDivElement, LoaderInputProps>((props, ref) => {
    const { loading, children, className, ...attributes } = props

    return (
        <div ref={ref} className={classNames(styles.container, className)} {...attributes}>
            {children}
            {loading && <LoadingSpinner inline={false} className={styles.spinner} />}
        </div>
    )
})
