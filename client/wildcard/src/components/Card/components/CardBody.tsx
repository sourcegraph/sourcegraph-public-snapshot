import classNames from 'classnames'
import React from 'react'

import styles from './CardBody.module.scss'

export const CardBody: React.FunctionComponent<React.HTMLAttributes<HTMLDivElement>> = ({
    children,
    className,
    ...attributes
}) => (
    <div className={classNames(className, styles.cardBody)} {...attributes}>
        {children}
    </div>
)
