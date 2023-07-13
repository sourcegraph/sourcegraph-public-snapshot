import { FC, useLayoutEffect, useMemo, useRef, useState } from 'react'

import styles from './Breadcrumbs.module.scss'
import classNames from 'classnames';
import { Button } from '../Button';

type Index = number
type Width = number

enum SegmentType {
    Common,
    Invisible,
    MoreButton
}

type Segment =
    { type: SegmentType.MoreButton } |
    { type: SegmentType.Invisible, value: string } |
    { type: SegmentType.Common, value: string }

interface BreadcrumbsProps {
    filename: string
    className?: string
}

export const Breadcrumbs: FC<BreadcrumbsProps> = props => {
    const { filename, className } = props

    const [hidedSegments, setHidedSegments] = useState<number[]>([])
    const rootElementRef = useRef<HTMLUListElement>(null)
    const segments = useMemo(() => filename.split('/'), [filename])

    useLayoutEffect(() => {
        if (!rootElementRef.current) {
            return
        }

        function fixItemsAppearance(width: number) {
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
                const segmentsToHide = []
                const middleElementIndex = Math.floor((segmentsElements.length - 1) / 2)

                while (totalWidth > width - 40) {
                    const elementToRemoveIndex = middleElementIndex + offset

                    totalWidth -= elementSizesMap[elementToRemoveIndex]
                    segmentsToHide.push(elementToRemoveIndex)

                    // Produce the sequence 0, 1, -1, 2, -2, 3, -3, ....
                    if (offset === 0) {
                        offset = 1
                    } else if (offset > 0) {
                        offset = offset * -1
                    } else {
                        offset = offset * -1 + 1
                    }
                }

                setHidedSegments(segmentsToHide)
            }
        }

        // Force initial fix item appearance synchronously to avoid flashes
        fixItemsAppearance(rootElementRef.current.getBoundingClientRect().width ?? 0)

        const resizeObserver = new ResizeObserver(entries => {
            const entry = entries[0]

            fixItemsAppearance(entry.contentRect.width)
        })

        resizeObserver.observe(rootElementRef.current)

        return () => resizeObserver.disconnect()
    }, [filename])

    const fixedSegments = useMemo<Segment[]>(() => {
        if (hidedSegments.length === 0) {
            return segments
        }

        const result: Segment[] = []

        for (const [index, segment] of segments.entries()) {
            if (hidedSegments.includes(index)) {
                result.push({
                    type: SegmentType.Invisible,
                    value: segment
                })
            } else {
                result.push({
                    type: SegmentType.Common,
                    value: segment
                })
            }
        }

        const firstInvisibleElement = result.findIndex(item => item.type === SegmentType.Invisible)

        if (firstInvisibleElement === -1) {
            result.splice(firstInvisibleElement, 0, { type: SegmentType.MoreButton })
        }

        return result
    }, [segments, hidedSegments])

    return (
        <ul ref={rootElementRef} className={classNames(styles.list, className)}>
            { segments.map((segment, index) =>
                <li
                    key={`${segment}-${index}`}
                    data-index={index}
                    className={classNames(styles.item, { [styles.itemHidden]: hidedSegments.includes(index)})}
                >
                    { getSegmentText(segments, index, segment) }
                </li>
            )}
        </ul>
    )
}

function getSegmentText(segments: string[], index: number, segment: string): string {
    return index !== segments.length - 1 ? `${segment}/` : segment
}

interface TruncatedItemsButton {
    segments: string[]
    truncatedSegments: Set<number>
}

const TruncatedItemsButton: FC<TruncatedItemsButton> = props => {
    const { segments, truncatedSegments} = props

    return (
        <Button variant='secondary' outline={true}>
            ...
        </Button>
    )
}
