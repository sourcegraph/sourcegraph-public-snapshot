import { type FC, type ReactNode, useLayoutEffect, useMemo, useRef, useState } from 'react'

import { mdiFolderOutline } from '@mdi/js'
import classNames from 'classnames'

import { Icon } from '../Icon'
import { Link } from '../Link'
import { Menu, MenuButton, MenuLink, MenuList } from '../Menu'

import styles from './Breadcrumbs.module.scss'

type Index = number
type Width = number

enum SegmentType {
    Common,
    Invisible,
    MoreButton,
}

interface CommonSegment {
    type: SegmentType.Common
    id: number
    value: string
}

interface InvisibleSegment {
    type: SegmentType.Invisible
    id: number
    value: string
}

interface MoreButtonSegment {
    type: SegmentType.MoreButton
    id: number
}

type Segment = CommonSegment | InvisibleSegment | MoreButtonSegment

const isInvisibleSegment = (segment: Segment): segment is InvisibleSegment => segment.type === SegmentType.Invisible
const isCommonSegment = (segment: Segment): segment is CommonSegment => segment.type === SegmentType.Common
const isMoreButtonSegment = (segment: Segment): segment is MoreButtonSegment => segment.type === SegmentType.MoreButton

// Static width value for the more width value, it's used below
// in items layout calculation in useLayoutEffect
const MORE_BUTTON_WIDTH = 40

interface BreadcrumbsProps {
    filename: string
    className?: string
    children?: ReactNode
    getSegmentLink: (segment: string, index: number, segments: string[]) => string
}

export const Breadcrumbs: FC<BreadcrumbsProps> = props => {
    const { filename, className, children, getSegmentLink } = props

    const [hiddenSegments, setHiddenSegments] = useState<Set<number>>(() => new Set())
    const rootElementRef = useRef<HTMLUListElement>(null)
    const segments = useMemo(() => filename.split('/'), [filename])

    useLayoutEffect(() => {
        if (!rootElementRef.current) {
            return
        }

        function fixItemsAppearance(width: number): void {
            // Base guards for root element and it's width
            if (!rootElementRef.current || width === 0) {
                return
            }

            // Measure segments elements sizes
            let totalWidth = 0
            const elementSizesMap: Record<Index, Width> = {}
            const extraContentElement =
                rootElementRef.current.querySelector<HTMLLIElement>('[data-type="extra-content"]')
            const segmentsElements = rootElementRef.current.querySelectorAll<HTMLLIElement>('[data-type="common"]')

            for (const element of segmentsElements) {
                const index = +(element.dataset?.index ?? 0)
                const width = element.getBoundingClientRect().width
                elementSizesMap[index] = width

                totalWidth += width
            }

            totalWidth += extraContentElement ? extraContentElement.getBoundingClientRect().width : 0

            // Elements overflow parent container, we should remove some elements in the middle
            // until all reset elements fit in the parent element
            if (totalWidth > width) {
                let offset = 0
                const segmentsToHide = []
                const middleElementIndex = Math.floor((segmentsElements.length - 1) / 2)

                while (totalWidth > width - MORE_BUTTON_WIDTH) {
                    const elementToRemoveIndex = middleElementIndex + offset

                    // Always render the last segment (truncation of the last segment is
                    // handled by CSS truncation
                    if (elementToRemoveIndex === segmentsElements.length - 1) {
                        break
                    }

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

                setHiddenSegments(new Set(segmentsToHide))
            } else {
                setHiddenSegments(new Set([]))
            }
        }

        // Force initial fix item appearance synchronously to avoid flashes
        fixItemsAppearance(rootElementRef.current.getBoundingClientRect().width ?? 0)

        const resizeObserver = new ResizeObserver(([entry]) => {
            fixItemsAppearance(entry.contentRect.width)
        })

        resizeObserver.observe(rootElementRef.current)

        return () => resizeObserver.disconnect()
    }, [filename])

    const fixedSegments = useMemo<Segment[]>(() => {
        const result = segments.map<Segment>((segment, index) => ({
            id: index,
            type: hiddenSegments.has(index) ? SegmentType.Invisible : SegmentType.Common,
            value: segment,
        }))

        const firstInvisibleElement = result.findIndex(item => item.type === SegmentType.Invisible)

        if (firstInvisibleElement !== -1) {
            result.splice(firstInvisibleElement, 0, {
                id: -1,
                type: SegmentType.MoreButton,
            })
        }

        return result
    }, [segments, hiddenSegments])

    const isOnlyFileNameVisible = fixedSegments.filter(isCommonSegment).length === 1

    return (
        <ul ref={rootElementRef} className={classNames(styles.list, className)}>
            {fixedSegments.map((segment, index) => (
                <li
                    key={segment.id}
                    data-index={segment.id}
                    data-type={isMoreButtonSegment(segment) ? 'internal' : 'common'}
                    className={classNames({
                        [styles.itemHidden]: isInvisibleSegment(segment),
                        [styles.itemWithButton]: isMoreButtonSegment(segment),
                        [styles.itemLast]: isOnlyFileNameVisible && index === fixedSegments.length - 1,
                        // Only for test purpose (see nav.test.ts integration test)
                        'test-breadcrumb-part-last': index === fixedSegments.length - 1,
                    })}
                >
                    {isMoreButtonSegment(segment) && (
                        <>
                            <TruncatedItemsButton
                                segments={segments}
                                truncatedSegments={fixedSegments.filter(isInvisibleSegment)}
                                getSegmentLink={getSegmentLink}
                            />
                            <Separator />
                        </>
                    )}

                    {!isMoreButtonSegment(segment) && (
                        <>
                            <Link to={getSegmentLink(segment.value, segment.id, segments)}>{segment.value}</Link>
                            {index !== fixedSegments.length - 1 && <Separator />}
                        </>
                    )}
                </li>
            ))}
            {children && <li data-type="extra-content">{children}</li>}
        </ul>
    )
}

const Separator: FC = () => <span className={styles.separator}>/</span>

interface TruncatedItemsButton {
    truncatedSegments: InvisibleSegment[]
    segments: string[]
    getSegmentLink: (segment: string, index: number, segments: string[]) => string
}

const TruncatedItemsButton: FC<TruncatedItemsButton> = props => {
    const { truncatedSegments, segments, getSegmentLink } = props

    return (
        <Menu>
            {({ isOpen }) => (
                <>
                    <MenuButton
                        variant="secondary"
                        outline={true}
                        className={classNames(styles.moreButton, { [styles.moreButtonActive]: isOpen })}
                    >
                        ...
                    </MenuButton>

                    <MenuList as="ul" className={styles.truncatedList}>
                        {truncatedSegments.map(segment => (
                            <MenuLink
                                key={segment.id}
                                as={Link}
                                to={getSegmentLink(segment.value, segment.id, segments)}
                                className={styles.truncatedListItem}
                            >
                                <Icon svgPath={mdiFolderOutline} aria-hidden={true} />
                                {segment.value}
                            </MenuLink>
                        ))}
                    </MenuList>
                </>
            )}
        </Menu>
    )
}
