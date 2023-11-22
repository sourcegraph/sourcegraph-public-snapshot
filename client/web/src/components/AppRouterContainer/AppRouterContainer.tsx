import React, { type HTMLAttributes, useRef } from 'react'

import classNames from 'classnames'

import { useScrollManager } from '../../hooks'

import styles from './AppRouterContainer.module.scss'

type AppRouterContainerProps = HTMLAttributes<HTMLDivElement>

export const AppRouterContainer: React.FunctionComponent<React.PropsWithChildren<AppRouterContainerProps>> = ({
    children,
    className,
    ...rest
}) => {
    const containerRef = useRef<HTMLElement>(null)
    useScrollManager('AppRouterContainer', containerRef)

    return (
        // Data Layout data attribute is used to get access to the main layout scroll element (the div element
        // below) from child levels in order to handle or react on this container scroll or other important events.
        <main
            ref={containerRef}
            data-layout={true}
            className={classNames(styles.appRouterContainer, className)}
            {...rest}
        >
            {children}
        </main>
    )
}
