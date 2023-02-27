import { FunctionComponent, HTMLAttributes, useRef } from 'react'

import classNames from 'classnames'

import { useElementObscuredArea } from '../../../../../hooks'

import styles from './ScrollBox.module.scss'

interface ScrollBoxProps extends HTMLAttributes<HTMLDivElement> {
    lazyMeasurements?: boolean
    className?: string
}

export const ScrollBox: FunctionComponent<ScrollBoxProps> = props => {
    const { lazyMeasurements, children, className, ...otherProps } = props

    const scrollRef = useRef<HTMLDivElement>(null)
    const obscuredArea = useElementObscuredArea(scrollRef, lazyMeasurements)

    const shutterClasses = [
        obscuredArea.top > 0 ? styles.rootWithTopFader : undefined,
        obscuredArea.bottom > 0 ? styles.rootWithBottomFader : undefined,
    ]

    return (
        <div {...otherProps} className={classNames(styles.root, className, ...shutterClasses)}>
            <div ref={scrollRef} className={styles.scrollContainer}>
                {children}
            </div>
        </div>
    )
}
