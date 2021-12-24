import classNames from 'classnames'
import React from 'react'

import styles from './CardSubtitle.module.scss'

export const CardSubtitle: React.FunctionComponent<React.HTMLAttributes<HTMLDivElement>> = ({
    children,
    className,
    ...attributes
}) => (
    <div className={classNames(className, styles.cardSubtitle)} {...attributes}>
        {children}
    </div>
)
