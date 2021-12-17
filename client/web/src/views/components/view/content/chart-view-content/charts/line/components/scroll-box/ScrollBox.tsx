import classNames from 'classnames'
import React, { useEffect, useState } from 'react'

import { observeResize } from '../../../../../../../../../util/dom'

import styles from './ScrollBox.module.scss'

/**
 * Mutates element (adds/removes css classes) based on scroll height and client height.
 */
function addShutterElementsToTarget(element: HTMLDivElement): void {
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

    // Catch element reference with useState to trigger elements update through useEffect
    const [scrollElement, setScrollElement] = useState<HTMLDivElement | null>()
    const [hasScroll, setHasScroll] = useState(false)

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

        const resizeSubscription = observeResize(scrollElement).subscribe(entry => {
            if (!entry) {
                return
            }

            const { target, contentRect } = entry

            setHasScroll(target.scrollHeight > contentRect.height)
            addShutterElementsToTarget(scrollElement)
        })

        return () => {
            scrollElement.removeEventListener('scroll', onScroll)
            resizeSubscription.unsubscribe()
        }
    }, [scrollElement])

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
