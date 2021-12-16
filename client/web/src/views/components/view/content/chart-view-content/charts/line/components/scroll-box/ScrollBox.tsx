import classNames from 'classnames'
import React, { useEffect, useState } from 'react'

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

const hasVerticalScroll = (element: HTMLElement): boolean => element.scrollHeight > element.clientHeight

interface ScrollBoxProps extends React.HTMLAttributes<HTMLDivElement> {
    className?: string
}

export const ScrollBox: React.FunctionComponent<ScrollBoxProps> = props => {
    const { children, className, ...otherProps } = props

    // Catch element reference with useState to trigger elements update through useEffect
    const [scrollElement, setScrollElement] = useState<HTMLDivElement | null>()

    useEffect(() => {
        if (!scrollElement) {
            return
        }

        // On mount initial call
        addShutterElementsToTarget(scrollElement)
    })

    useEffect(() => {
        if (!scrollElement) {
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

        scrollElement.addEventListener('scroll', onScroll)

        return () => scrollElement.removeEventListener('scroll', onScroll)
    }, [scrollElement])

    const hasScroll = scrollElement ? hasVerticalScroll(scrollElement) : false

    return (
        <div {...otherProps} ref={setScrollElement} className={classNames(styles.root, className)}>
            {hasScroll && (
                <>
                    <div className={classNames(styles.fader, styles.faderTop)} />
                    <div className={classNames(styles.fader, styles.faderBottom)} />
                </>
            )}

            <div className={classNames({ [styles.scrollbox]: hasScroll })}>{children}</div>
        </div>
    )
}
