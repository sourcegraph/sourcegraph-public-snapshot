import { HTMLAttributes } from 'react'
import * as React from 'react'

import classNames from 'classnames'

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
