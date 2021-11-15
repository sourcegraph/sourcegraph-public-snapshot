import classNames from 'classnames'
import React, { HTMLAttributes } from 'react'

import styles from './AppRouterContainer.module.scss'

type AppRouterContainerProps = HTMLAttributes<HTMLDivElement>

export const AppRouterContainer: React.FunctionComponent<AppRouterContainerProps> = ({
    children,
    className,
    ...rest
}) => (
    <div className={classNames(styles.appRouterContainer, className)} {...rest}>
        {children}
    </div>
)
