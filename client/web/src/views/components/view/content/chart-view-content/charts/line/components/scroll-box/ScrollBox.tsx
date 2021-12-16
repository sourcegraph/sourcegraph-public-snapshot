import classNames from 'classnames'
import React, { useEffect, useRef } from 'react'

import styles from './ScrollBox.module.scss'

const addShutterElementsToTarget = (element: HTMLDivElement): void => {
    const { scrollTop, scrollHeight, offsetHeight } = element

    if (scrollTop === 0) {
        element.classList.remove(styles.rootWithTopFader)
    } else {
        element.classList.add(styles.rootWithTopFader)
    }

    if (offsetHeight + scrollTop === scrollHeight) {
        element.classList.remove(styles.rootWithBottomFader)
    } else {
        element.classList.add(styles.rootWithBottomFader)
    }
}

interface ScrollBoxProps extends React.HTMLAttributes<HTMLDivElement> {
    className?: string
}

export const ScrollBox: React.FunctionComponent<ScrollBoxProps> = props => {
    const { children, className, ...otherProps } = props

    const scrollBoxReference = useRef<HTMLDivElement>(null)

    useEffect(() => {
        const scrollBoxElement = scrollBoxReference.current

        if (!scrollBoxElement) {
            return
        }

        // On mount initial call
        addShutterElementsToTarget(scrollBoxElement)
    })

    useEffect(() => {
        const scrollBoxElement = scrollBoxReference.current

        if (!scrollBoxElement) {
            return
        }

        function onScroll(event: Event): void {
            if (!event.target) {
                return
            }

            addShutterElementsToTarget(event.target as HTMLDivElement)

            event.stopPropagation()
            event.preventDefault()
        }

        scrollBoxElement.addEventListener('scroll', onScroll)

        return () => scrollBoxElement.removeEventListener('scroll', onScroll)
    }, [])

    return (
        <div {...otherProps} ref={scrollBoxReference} className={classNames(styles.root, className)}>
            <div className={classNames(styles.fader, styles.faderTop)} />
            <div className={classNames(styles.fader, styles.faderBottom)} />

            <div className={styles.scrollbox}>{children}</div>
        </div>
    )
}
