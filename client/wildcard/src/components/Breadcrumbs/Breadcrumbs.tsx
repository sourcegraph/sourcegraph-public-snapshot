import { FC, useLayoutEffect, useMemo, useRef, useState } from 'react'

import styles from './Breadcrumbs.module.scss'
import classNames from 'classnames';
import useResizeObserver from 'use-resize-observer';

type Index = number
type Width = number

interface BreadcrumbsProps {
    filename: string
    className?: string
}

export const Breadcrumbs: FC<BreadcrumbsProps> = props => {
    const { filename, className } = props

    const rootElementRef = useRef<HTMLUListElement>(null)
    const { width = 0 } = useResizeObserver({ ref: rootElementRef })

    const segmentsRef = useRef(() => new Set())
    const segments = useMemo(() => filename.split('/'), [filename])

    useLayoutEffect(() => {
        // Base guards for root element and it's width
        if (!rootElementRef.current || width === 0) {
            return
        }

        // Measure segments elements sizes
        let totalWidth = 0
        const elementSizesMap: Record<Index, Width> = {}
        const segmentsElements = rootElementRef.current.querySelectorAll('li')

        for (const element of segmentsElements) {
            const index = +(element.dataset?.index ?? 0)
            const width = element.getBoundingClientRect().width
            elementSizesMap[index] = width

            totalWidth += width
        }

        // Elements overflow parent container, we should remove some elements in the middle
        // until all reset elements fit in the parent element
        if (totalWidth > width) {
            let offset: number = 0
            const middleElementIndex = Math.floor(segmentsElements.length - 1 / 2)

            while (totalWidth > width - 40) {
                const elementToRemoveIndex = middleElementIndex + offset

                offset = offset === 0 ? 1 : offset - (-offset)
            }
        }

    }, [filename, width])

    return (
        <ul ref={rootElementRef} className={classNames(styles.list, className)}>
            { segments.map((segment, index) =>
                <li key={`${segment}-${index}`} data-index={index}>
                    { index !== segments.length - 1 ? `${segment}/` : segment}
                </li>
            )}
        </ul>
    )
}
