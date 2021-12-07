import classNames from 'classnames'
import React, { useEffect, useRef } from 'react'

import styles from './ScrollBox.module.scss'

const addShutterElementsToTarget = (element: HTMLDivElement, fader: HTMLDivElement): void => {
    const { scrollTop, scrollHeight, offsetHeight } = element

    if (scrollTop === 0) {
        fader?.classList.remove(styles.faderHasTopScroll)
    } else {
        fader?.classList.add(styles.faderHasTopScroll)
    }

    if (offsetHeight + scrollTop === scrollHeight) {
        fader?.classList.remove(styles.faderHasBottomScroll)
    } else {
        fader?.classList.add(styles.faderHasBottomScroll)
    }
}

interface ScrollBoxProps extends React.HTMLAttributes<HTMLDivElement> {
    scrollEnabled?: boolean
    as?: React.ElementType
    rootClassName?: string
}

export const ScrollBox: React.FunctionComponent<ScrollBoxProps> = props => {
    const { children, as: Component = 'div', scrollEnabled = true, rootClassName, ...otherProps } = props

    const scrollBoxReference = useRef<HTMLDivElement>(null)
    const faderReference = useRef<HTMLDivElement>(null)

    useEffect(() => {
        const fader = faderReference.current
        const scrollBoxElement = scrollBoxReference.current

        if (!scrollBoxElement || !fader || !scrollEnabled) {
            return
        }

        // On mount initial call
        addShutterElementsToTarget(scrollBoxElement, fader)

        function onScroll(event: Event): void {
            if (!event.target) {
                return
            }

            addShutterElementsToTarget(event.target as HTMLDivElement, fader as HTMLDivElement)

            event.stopPropagation()
            event.preventDefault()
        }

        scrollBoxElement.addEventListener('scroll', onScroll)

        return () => scrollBoxElement.removeEventListener('scroll', onScroll)
    }, [scrollEnabled])

    // If block doesn't have scroll content we render simple component without additional
    // shutter elements for scroll visual effects
    if (!scrollEnabled) {
        return (
            <Component {...otherProps} className={classNames(otherProps.className, rootClassName)}>
                {children}
            </Component>
        )
    }

    return (
        <div {...otherProps} className={classNames(styles.root, rootClassName)}>
            <div ref={faderReference} className={styles.fader} />
            <Component ref={scrollBoxReference} className={classNames(otherProps.className, styles.scrollbox)}>
                {children}
            </Component>
        </div>
    )
}
