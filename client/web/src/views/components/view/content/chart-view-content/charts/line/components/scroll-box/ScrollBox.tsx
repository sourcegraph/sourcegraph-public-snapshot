import classNames from 'classnames'
import React, { useLayoutEffect, useRef, useState } from 'react'

import { isFirefox } from '@sourcegraph/shared/src/util/browserDetection'

import { observeResize } from '../../../../../../../../../util/dom'

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
        ? target.scrollHeight > Math.round(target.clientHeight + target.clientHeight / 100)
        : target.scrollHeight > target.clientHeight

    target.style.overflow = ''

    return hasScroll
}

interface ScrollBoxProps extends React.HTMLAttributes<HTMLDivElement> {
    className?: string
}

export const ScrollBox: React.FunctionComponent<ScrollBoxProps> = props => {
    const { children, className, ...otherProps } = props

    const parentElementReference = useRef<HTMLDivElement>(null)
    const [hasScroll, setHasScroll] = useState(false)

    useLayoutEffect(() => {
        const parentElement = parentElementReference.current
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

            setHasScroll(hasElementScroll(target))
            addShutterElementsToTarget(parentElement, parentElement)
        })

        return () => {
            parentElement.removeEventListener('scroll', onScroll)
            resizeSubscription.unsubscribe()
        }
    }, [])

    return (
        <div
            {...otherProps}
            ref={parentElementReference}
            className={classNames(styles.root, className, { [styles.rootWithScroll]: hasScroll })}
        >
            {hasScroll && (
                <>
                    <div className={classNames(styles.fader, styles.faderTop)} />
                    <div className={classNames(styles.fader, styles.faderBottom)} />
                </>
            )}

            <div className={classNames({ [styles.scrollboxWithScroll]: hasScroll })}>{children}</div>
        </div>
    )
}
