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
        {/*
            Data Layout data attribute is used to get access to the main layout scroll element (the div element below).
            from child levels in order to handle or react on this container scroll or other important events.
         */}
        <div data-layout={true} className={classNames(styles.appRouterContainer, className)} {...rest}>
            {children}
        </div>
    </ElementScroller>
)
