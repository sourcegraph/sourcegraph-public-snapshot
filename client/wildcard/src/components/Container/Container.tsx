import React, { ReactNode } from 'react'

import classNames from 'classnames'

import styles from './Container.module.scss'

interface Props {
    className?: string
    children?: ReactNode
}

/** A container wrapper. Used for grouping content together. */
export const Container: React.FunctionComponent<Props> = ({ children, className }) => (
    <div className={classNames(styles.container, className)}>{children}</div>
)
