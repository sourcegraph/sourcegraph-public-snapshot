import classNames from 'classnames'
import React from 'react'

import styles from './CardText.module.scss'

export const CardText: React.FunctionComponent<React.HTMLAttributes<HTMLDivElement>> = ({
    children,
    className,
    ...attributes
}) => (
    <p className={classNames(className, styles.cardText)} {...attributes}>
        {children}
    </p>
)
