import classNames from 'classnames'
import React, { HTMLAttributes } from 'react'

import styles from './ListItemContainer.module.scss'

type ListItemContainerProps = HTMLAttributes<HTMLLIElement>

export const ListItemContainer: React.FunctionComponent<ListItemContainerProps> = ({ children, ...rest }) => (
    <li className={classNames('list-group-item', styles.container)} {...rest}>
        {children}
    </li>
)
