import React, { useEffect, useState } from 'react'

import classNames from 'classnames'

import { isFirefox, observeResize } from '@sourcegraph/common'

import styles from './ScrollBox.module.scss'

/**
 * Mutates element (adds/removes css classes) based on scroll height and client height.
 */
function addShutterElementsToTarget(parentElement: HTMLDivElement, scrollElement: HTMLDivElement): void {
    const { scrollTop, scrollHeight, offsetHeight } = scrollElement

    if (scrollTop === 0) {
        parentElement.classList.remove(styles.rootWithTopFader)
    } else {
        parentElement.classList.add(styles.rootWithTopFader)
    }

    if (offsetHeight + scrollTop === scrollHeight) {
        parentElement.classList.remove(styles.rootWithBottomFader)
    } else {
        parentElement.classList.add(styles.rootWithBottomFader)
    }
}

function hasElementScroll(target: HTMLElement): boolean {
    target.style.overflow = 'scroll'

    const hasScroll = isFirefox()
        ? // For some reason in Firefox it's possible to get a wrong scrollHeight with 1% ~ 1px
          // static error. To avoid this "fake" scroll we take into calculation a static error
          // value which equals to 1% of scrollable container height.
          target.scrollHeight > Math.round(target.clientHeight + target.clientHeight / 100)
        : target.scrollHeight > target.clientHeight

    target.style.overflow = ''

    return hasScroll
}

interface ScrollBoxProps extends React.HTMLAttributes<HTMLDivElement> {
    className?: string
}

export const ScrollBox: React.FunctionComponent<React.PropsWithChildren<ScrollBoxProps>> = props => {
    const { children, className, ...otherProps } = props

    const [parentElement, setParentElement] = useState<HTMLDivElement | null>()
    const [hasScroll, setHasScroll] = useState(false)

    useEffect(() => {
        if (!parentElement) {
            return
        }

        function onScroll(event: Event): void {
            if (!event.target || !parentElement) {
                return
            }

            addShutterElementsToTarget(parentElement, event.target as HTMLDivElement)

            event.stopPropagation()
            event.preventDefault()
        }

        parentElement.addEventListener('scroll', onScroll)

        const resizeSubscription = observeResize(parentElement).subscribe(entry => {
            if (!entry) {
                return
            }

            const target = entry.target as HTMLElement

            // Delay overflow content measurements while browser renders
            // parent content
            requestAnimationFrame(() => {
                setHasScroll(hasElementScroll(target))
                addShutterElementsToTarget(parentElement, parentElement)
            })
        })

        return () => {
            parentElement.removeEventListener('scroll', onScroll)
            resizeSubscription.unsubscribe()
        }
    }, [parentElement])

    return (
        <div
            {...otherProps}
            ref={setParentElement}
            className={classNames(styles.root, className, { [styles.rootWithScroll]: hasScroll })}
        >
            {hasScroll && (
                <>
                    <div className={classNames(styles.fader, styles.faderTop)} />
                    <div className={classNames(styles.fader, styles.faderBottom)} />
                </>
            )}

            {children}
        </div>
    )
}
