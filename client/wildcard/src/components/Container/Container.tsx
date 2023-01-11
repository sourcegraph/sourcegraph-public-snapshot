import React from 'react'

import classNames from 'classnames'

import styles from './Container.module.scss'

interface Props extends React.HTMLAttributes<HTMLDivElement> {
    className?: string
}

/** A container wrapper. Used for grouping content together. */
export const Container: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    children,
    className,
    ...props
}) => (
    <div className={classNames(styles.container, className)} {...props}>
        {children}
    </div>
)
