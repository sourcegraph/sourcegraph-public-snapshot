import classNames from 'classnames'
import React from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

import styles from './LoaderInput.module.scss'

/** Takes loading prop, input component as child */

interface Props {
    loading: boolean
    children: React.ReactNode
    className?: string
}

export const LoaderInput: React.FunctionComponent<Props> = ({ loading, children, className }) => (
    <div className={classNames(styles.container, className)}>
        {children}
        {loading && <LoadingSpinner className={styles.spinner} />}
    </div>
)
