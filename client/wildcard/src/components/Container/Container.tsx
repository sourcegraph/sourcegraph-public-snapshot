import classNames from 'classnames'
import React from 'react'

import styles from './Container.module.scss'

interface Props {
    className?: string
}

/** A container wrapper. Used for grouping content together. */
export const Container: React.FunctionComponent<Props> = ({ children, className }) => (
    <div className={classNames(styles.container, className)}>{children}</div>
)
