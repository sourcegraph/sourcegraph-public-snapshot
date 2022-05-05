import React, { HTMLAttributes } from 'react'

import classNames from 'classnames'

import styles from './RepositoryNodeContainer.module.scss'

type RepositoryNodeContainerProps = {
    as: React.ElementType<HTMLAttributes<HTMLTableRowElement | HTMLTableCellElement>>
} & HTMLAttributes<HTMLTableRowElement | HTMLTableCellElement>

export const RepositoryNodeContainer: React.FunctionComponent<
    React.PropsWithChildren<RepositoryNodeContainerProps>
> = ({ children, className, as: Component, ...rest }) => (
    <Component className={classNames(className, styles.repositoryNode)} {...rest}>
        {children}
    </Component>
)
