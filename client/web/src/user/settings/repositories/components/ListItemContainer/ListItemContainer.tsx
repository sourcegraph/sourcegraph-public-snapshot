import React, { HTMLAttributes } from 'react'

import classNames from 'classnames'

import styles from './ListItemContainer.module.scss'

type ListItemContainerProps = HTMLAttributes<HTMLLIElement>

export const ListItemContainer: React.FunctionComponent<React.PropsWithChildren<ListItemContainerProps>> = ({
    children,
    ...rest
}) => (
    <li className={classNames('list-group-item', styles.container)} {...rest}>
        {children}
    </li>
)
