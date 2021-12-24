import classNames from 'classnames'
import React from 'react'

import styles from './CardHeader.module.scss'

export const CardHeader: React.FunctionComponent<React.HTMLAttributes<HTMLDivElement>> = ({
    children,
    className,
    ...attributes
}) => (
    <div className={classNames(className, styles.cardHeader)} {...attributes}>
        {children}
    </div>
)
