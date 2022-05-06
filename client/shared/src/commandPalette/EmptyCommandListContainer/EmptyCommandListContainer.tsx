import React, { HTMLAttributes } from 'react'

import classNames from 'classnames'

import styles from './EmptyCommandListContainer.module.scss'

type EmptyCommandListContainerProps = HTMLAttributes<HTMLDivElement>

export const EmptyCommandListContainer: React.FunctionComponent<
    React.PropsWithChildren<EmptyCommandListContainerProps>
> = ({ className, children, ...rest }) => (
    <div className={classNames(styles.emptyCommandList, className)} {...rest}>
        {children}
    </div>
)
