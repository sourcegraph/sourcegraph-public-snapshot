import classNames from 'classnames'
import React, { HTMLAttributes } from 'react'

import styles from './EmptyCommandListContainer.module.scss'

type EmptyCommandListContainerProps = HTMLAttributes<HTMLDivElement>

export const EmptyCommandListContainer: React.FunctionComponent<EmptyCommandListContainerProps> = ({
    className,
    children,
    ...rest
}) => (
    <div className={classNames(styles.emptyCommandList, className)} {...rest}>
        {children}
    </div>
)
