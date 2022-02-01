import classNames from 'classnames'
import React, { HTMLAttributes } from 'react'
import { ElementScroller } from 'react-scroll-manager'

import styles from './AppRouterContainer.module.scss'

type AppRouterContainerProps = HTMLAttributes<HTMLDivElement>

export const AppRouterContainer: React.FunctionComponent<AppRouterContainerProps> = ({
    children,
    className,
    ...rest
}) => (
    <ElementScroller scrollKey="app-router-container">
        <div className={classNames(styles.appRouterContainer, className)} {...rest}>
            {children}
        </div>
    </ElementScroller>
)
