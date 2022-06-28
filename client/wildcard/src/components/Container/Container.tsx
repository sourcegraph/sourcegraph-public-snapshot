import React from 'react'

import classNames from 'classnames'

import styles from './Container.module.scss'

interface Props {
    className?: string
}

/** A container wrapper. Used for grouping content together. */
export const Container: React.FunctionComponent<React.PropsWithChildren<Props>> = ({ children, className }) => (
    <div className={classNames(styles.container, className)}>{children}</div>
)
