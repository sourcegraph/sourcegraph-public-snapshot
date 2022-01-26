import classNames from 'classnames'
import React, { HTMLAttributes } from 'react'

import styles from './ConnectionPopoverNode.module.scss'

type ConnectionPopoverNodeProps = HTMLAttributes<HTMLLIElement>

export const ConnectionPopoverNode: React.FunctionComponent<ConnectionPopoverNodeProps> = ({
    className,
    children,
    ...rest
}) => (
    <li className={classNames(styles.connectionPopoverNode, className)} {...rest}>
        {children}
    </li>
)
