import React, { useEffect, useRef } from 'react'

import classNames from 'classnames'

import styles from './DashboardHeader.module.scss'

interface DashboardHeaderProps extends React.HTMLAttributes<HTMLElement> {}

export const DashboardHeader: React.FunctionComponent<React.PropsWithChildren<DashboardHeaderProps>> = props => {
    const { children } = props
    const reference = useRef(null)

    useEffect(() => {
        const element = reference.current

        if (!element) {
            return
        }

        // Here we trying to get closest scroll layout element
        // In our case it should be AppRouterContainer div element.
        // TODO: Put AppRouterContainer element in the global state to avoid bad querySelector calls
        const root = document.querySelector('[data-layout]')

        // Based on the solution from the following stackoverflow thread
        // https://stackoverflow.com/questions/16302483/event-to-detect-when-positionsticky-is-triggered
        const observer = new IntersectionObserver(
            ([entry]) => entry.target.classList.toggle(styles.headerStuck, entry.intersectionRatio < 1),
            { root, rootMargin: '-1px 0px 0px 0px', threshold: [1] }
        )

        observer.observe(element)

        return () => observer.disconnect()
    }, [])

    return (
        <header ref={reference} {...props} className={classNames(props.className, styles.header)}>
            {children}
        </header>
    )
}
