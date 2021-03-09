import React from 'react'
import classNames from 'classnames'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import useResizeObserver from 'use-resize-observer'
import { useDebounce } from '../hooks'

type Page = '...' | number

interface PageButton extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    active?: boolean
}

const PageButton: React.FunctionComponent<PageButton> = ({ children, active, ...props }) => (
    <li>
        <button
            type="button"
            className={classNames('btn', 'mx-1', 'page-selector__button', active && 'btn-primary')}
            {...props}
        >
            {children}
        </button>
    </li>
)

const range = (start: number, end: number) => {
    const length = end - start + 1
    return Array.from({ length }, (_, i) => start + i)
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

interface Props {
    currentPage: number
    onPageChange: (page: number) => void
    maxPages: number
    className?: string
}

export const PageSelector: React.FunctionComponent<Props> = ({ currentPage, onPageChange, maxPages, className }) => {
    const { ref, width } = useDebounce(useResizeObserver(), 50)
    const shouldShrink = width !== undefined && width < 576
    const pages = getPages({ currentPage, maxPages, shrink: shouldShrink })

    return (
        <nav>
            <ul ref={ref} className={classNames('page-selector', className)}>
                <PageButton disabled={currentPage === 1} onClick={() => onPageChange(currentPage - 1)}>
                    <ChevronLeftIcon className="icon-inline" aria-label="Previous" aria-hidden={!shouldShrink} />
                    <span className={classNames(shouldShrink && 'd-none')}>Previous</span>
                </PageButton>
                {pages.map((page, index) => (
                    <PageButton
                        key={index}
                        disabled={page === '...'}
                        onClick={() => page !== '...' && onPageChange(page)}
                        active={currentPage === page}
                    >
                        {page}
                    </PageButton>
                ))}
                <PageButton disabled={currentPage === maxPages} onClick={() => onPageChange(currentPage + 1)}>
                    <span className={classNames(shouldShrink && 'd-none')}>Next</span>
                    <ChevronRightIcon className="icon-inline" aria-label="Next" aria-hidden={!shouldShrink} />
                </PageButton>
            </ul>
        </nav>
    )
}
