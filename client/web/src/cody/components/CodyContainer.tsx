import React from 'react'

import classNames from 'classnames'

import styles from './CodyContainer.module.scss'

interface CodyContainerProps extends React.HTMLAttributes<HTMLDivElement> {
    className?: string
    children: React.ReactNode
}

export const CodyContainer: React.FunctionComponent<CodyContainerProps> = ({ className, children, ...props }) => {
    const containerClassName = classNames(styles.container, className)

    return (
        // eslint-disable-next-line no-restricted-syntax
        <div className={containerClassName} {...props}>
            {children}
        </div>
    )
}
