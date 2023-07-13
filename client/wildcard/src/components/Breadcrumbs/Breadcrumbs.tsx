import { FC, ReactNode, useLayoutEffect, useMemo, useRef, useState } from 'react'

import { mdiFolderOutline } from '@mdi/js'
import classNames from 'classnames'

import { Icon } from '../Icon'
import { Menu, MenuButton, MenuItem, MenuList } from '../Menu'

import styles from './Breadcrumbs.module.scss'

type Index = number
type Width = number

enum SegmentType {
    Common,
    Invisible,
    MoreButton,
}

interface MoreButtonSegment {
    type: SegmentType.MoreButton
    id: string
}
interface InvisibleSegment {
    type: SegmentType.Invisible
    id: string
    value: string
}
interface CommonSegment {
    type: SegmentType.Common
    id: string
    value: string
}

type Segment = CommonSegment | InvisibleSegment | MoreButtonSegment

const isInvisibleSegment = (segment: Segment): segment is InvisibleSegment => segment.type === SegmentType.Invisible
const isMoreButtonSegment = (segment: Segment): segment is MoreButtonSegment => segment.type === SegmentType.MoreButton

// Static width value for the more width value, it's used below
// in items layout calculation in useLayoutEffect
const MORE_BUTTON_WIDTH = 30

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
            const segmentsElements = rootElementRef.current.querySelectorAll<HTMLLIElement>('[data-type="common"]')

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

                while (totalWidth > width - MORE_BUTTON_WIDTH) {
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
            } else {
                setHidedSegments([])
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
            return segments.map((segment, index) => ({
                id: `${index}`,
                type: SegmentType.Common,
                value: segment,
            }))
        }

        const result: Segment[] = []

        for (const [index, segment] of segments.entries()) {
            if (hidedSegments.includes(index)) {
                result.push({
                    id: `${index}`,
                    type: SegmentType.Invisible,
                    value: segment,
                })
            } else {
                result.push({
                    id: `${index}`,
                    type: SegmentType.Common,
                    value: segment,
                })
            }
        }

        const firstInvisibleElement = result.findIndex(item => item.type === SegmentType.Invisible)

        if (firstInvisibleElement !== -1) {
            result.splice(firstInvisibleElement, 0, {
                id: 'more-items-button',
                type: SegmentType.MoreButton,
            })
        }

        return result
    }, [segments, hidedSegments])

    return (
        <ul ref={rootElementRef} className={classNames(styles.list, className)}>
            {fixedSegments.map((segment, index) => (
                <li
                    key={segment.id}
                    data-index={segment.id}
                    data-type={isMoreButtonSegment(segment) ? 'internal' : 'common'}
                    className={classNames(styles.item, {
                        [styles.itemHidden]: isInvisibleSegment(segment),
                        [styles.itemWithButton]: isMoreButtonSegment(segment),
                    })}
                >
                    {isMoreButtonSegment(segment) && (
                        <TruncatedItemsButton truncatedSegments={fixedSegments.filter(isInvisibleSegment)} />
                    )}
                    {getSegmentText(segment, index, fixedSegments)}
                </li>
            ))}
        </ul>
    )
}

function getSegmentText(segment: Segment, index: number, segments: Segment[]): ReactNode {
    const segmentText = !isMoreButtonSegment(segment) ? segment.value : ''

    return index !== segments.length - 1 ? (
        <>
            {segmentText}
            <span className={styles.separator}>/</span>
        </>
    ) : (
        segmentText
    )
}

interface TruncatedItemsButton {
    truncatedSegments: InvisibleSegment[]
}

const TruncatedItemsButton: FC<TruncatedItemsButton> = props => {
    const { truncatedSegments } = props

    return (
        <Menu>
            {({ isOpen }) => (
                <>
                    <MenuButton
                        variant="secondary"
                        outline={true}
                        className={classNames(styles.moreButton, { [styles.moreButtonActive]: isOpen })}
                    >
                        ... <span aria-hidden={true}>▾</span>
                    </MenuButton>

                    <MenuList as="ul" className={styles.truncatedList}>
                        {truncatedSegments.map(segment => (
                            <MenuItem key={segment.id} className={styles.truncatedListItem}>
                                <Icon svgPath={mdiFolderOutline} aria-hidden={true} />
                                {segment.value}
                            </MenuItem>
                        ))}
                    </MenuList>
                </>
            )}
        </Menu>
    )
}
