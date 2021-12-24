import classNames from 'classnames'
import React from 'react'

import styles from './CardList.module.scss'

export const CardList: React.FunctionComponent<React.HTMLAttributes<HTMLDivElement>> = ({
    children,
    className,
    ...attributes
}) => (
    <div className={classNames(className, styles.listGroup)} {...attributes}>
        {children}
    </div>
)
