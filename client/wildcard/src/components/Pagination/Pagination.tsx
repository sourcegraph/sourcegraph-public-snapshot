import React, { useCallback } from 'react'
import classNames from 'classnames'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import useResizeObserver from 'use-resize-observer/polyfilled'
import { useDebounce } from '@sourcegraph/wildcard'

type Page = '...' | number

interface PageButton extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    active?: boolean
}

const PageButton: React.FunctionComponent<PageButton> = ({ children, active, ...props }) => {
    const refocus = useCallback((button: HTMLButtonElement | null) => active && button?.focus(), [active])

    return (
        <li>
            <button
                type="button"
                className={classNames('btn', 'mx-1', 'pagination__button', active && 'btn-primary')}
                aria-current={active}
                ref={refocus}
                {...props}
            >
                {children}
            </button>
        </li>
    )
}

const range = (start: number, end: number): number[] => {
    const length = end - start + 1
    return Array.from({ length }, (value, index) => start + index)
}

const getMiddlePages = ({ currentPage, maxPages, shrink }: GetPages): Page[] => {
    const minimumSiblingRange = shrink ? 2 : 4
    const maximumSiblingRange = shrink ? 3 : 5
    const minimumBoundary = 2
    const maximumBoundary = maxPages - 1

    const siblingsStart = Math.max(
        Math.min(
            currentPage - minimumSiblingRange / 2,
            maxPages - maximumSiblingRange // Extend range when page is high
        ),
        minimumBoundary
    )

    const siblingsEnd = Math.min(
        Math.max(
            currentPage + minimumSiblingRange / 2,
            maximumSiblingRange + 1 // Extend range when page is low
        ),
        maximumBoundary
    )

    const middle: Page[] = range(siblingsStart, siblingsEnd)

    if (middle[0] !== minimumBoundary) {
        middle[0] = '...'
    }

    if (middle[middle.length - 1] !== maximumBoundary) {
        middle[middle.length - 1] = '...'
    }

    return middle
}

interface GetPages extends Pick<Props, 'currentPage' | 'maxPages'> {
    shrink: boolean
}

const getPages = ({ currentPage, maxPages, shrink }: GetPages): Page[] => {
    const pages: Page[] = [1]

    if (maxPages > 1) {
        pages.push(maxPages)
    }

    if (maxPages > 2) {
        pages.splice(1, 0, ...getMiddlePages({ currentPage, maxPages, shrink }))
    }

    return pages
}

export interface Props {
    currentPage: number
    onPageChange: (page: number) => void
    maxPages: number
    className?: string
}

const validatePositiveInteger = (name: string, value: number): string[] => {
    if (value > 0) {
        return []
    }
    return [`${name} must have a value greater than 0`]
}

const validateProps = (props: Props): void => {
    const errors = [
        ...validatePositiveInteger('maxPages', props.maxPages),
        ...validatePositiveInteger('currentPage', props.currentPage),
    ]

    if (props.currentPage > props.maxPages) {
        errors.push('currentPage must be not be greater than maxPages')
    }

    if (errors.length > 0) {
        throw new Error(`Invalid configuration:\n${errors.map(error => `— ${error}`).join('\n')}`)
    }
}

export const Pagination: React.FunctionComponent<Props> = props => {
    validateProps(props)

    const { maxPages, currentPage, className, onPageChange } = props
    const { ref, width } = useDebounce(useResizeObserver(), 100)
    const shouldShrink = width !== undefined && width <= 576
    const pages = getPages({ currentPage, maxPages, shrink: shouldShrink })

    const goBack = useCallback(() => onPageChange(currentPage - 1), [onPageChange, currentPage])
    const goForward = useCallback(() => onPageChange(currentPage + 1), [onPageChange, currentPage])
    const handleClick = useCallback(
        (page: Page) => (event: React.MouseEvent<HTMLButtonElement, MouseEvent>) => {
            if (page === '...') {
                return
            }

            // Blur the clicked page, as it will now shift and refocus
            event.currentTarget.blur()

            return onPageChange(page)
        },
        [onPageChange]
    )

    return (
        <nav>
            <ul ref={ref} className={classNames('pagination', className)}>
                <PageButton disabled={currentPage === 1} onClick={goBack} aria-label="Previous page">
                    <ChevronLeftIcon className="icon-inline" aria-label="Previous" aria-hidden={!shouldShrink} />
                    <span className={classNames(shouldShrink && 'd-none')}>Previous</span>
                </PageButton>
                {pages.map((page, index) => {
                    const isConnector = page === '...'
                    return (
                        <PageButton
                            key={index}
                            disabled={isConnector}
                            aria-hidden={isConnector}
                            aria-label={`Go to page ${page}`}
                            onClick={handleClick(page)}
                            active={currentPage === page}
                        >
                            {page}
                        </PageButton>
                    )
                })}
                <PageButton disabled={currentPage === maxPages} onClick={goForward} aria-label="Next page">
                    <span className={classNames(shouldShrink && 'd-none')}>Next</span>
                    <ChevronRightIcon className="icon-inline" aria-label="Next" aria-hidden={!shouldShrink} />
                </PageButton>
            </ul>
        </nav>
    )
}
