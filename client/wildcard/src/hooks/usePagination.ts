import { range } from 'lodash'
import { useCallback } from 'react'
import { useControlledState } from './useControlledState'

type Page = '...' | number

interface BasePaginationItem
    extends Pick<
        React.ButtonHTMLAttributes<HTMLButtonElement>,
        'onClick' | 'aria-current' | 'disabled' | 'aria-label'
    > {
    active: boolean
}

interface PreviousPaginationItem extends BasePaginationItem {
    type: 'previous'
    active: false
    content: 'Previous'
    'aria-label': 'Go to previous page'
}

interface NextPaginationItem extends BasePaginationItem {
    type: 'next'
    active: false
    content: 'Next'
    'aria-label': 'Go to next page'
}

interface ElipsisPaginationItem extends BasePaginationItem {
    type: 'elipsis'
    active: false
    disabled: true
    content: '...'
}

interface PagePaginationItem extends BasePaginationItem {
    type: 'page'
    content: number
    disabled: false
}

type PaginationItem = PreviousPaginationItem | NextPaginationItem | ElipsisPaginationItem | PagePaginationItem

const smallest = ({ values, minimum }: { values: number[]; minimum: number }): number =>
    Math.max(Math.min(...values), minimum)
const highest = ({ values, maximum }: { values: number[]; maximum: number }): number =>
    Math.min(Math.max(...values), maximum)

interface UsePaginationParameters {
    currentPage: number
    totalPages: number
    maxDisplayed: number
    onChange?: (page: number) => void
}

const getPages = ({ currentPage, totalPages, maxDisplayed }: UsePaginationParameters): Page[] => {
    const pages: Page[] = [1]

    if (totalPages > 1) {
        pages.push(totalPages)
    }

    if (totalPages === 2) {
        return pages
    }

    const minDisplayed = maxDisplayed - 1
    const minimumBoundary = 2
    const maximumBoundary = totalPages - 1

    const siblingsStart = smallest({
        values: [currentPage - minDisplayed / 2, totalPages - maxDisplayed],
        minimum: minimumBoundary,
    })
    const siblingsEnd = highest({
        values: [currentPage + minDisplayed / 2, maxDisplayed + 1],
        maximum: maximumBoundary,
    })

    const middle: Page[] = range(siblingsStart, siblingsEnd + 1)

    if (middle[0] !== minimumBoundary) {
        middle[0] = '...'
    }

    if (middle[middle.length - 1] !== maximumBoundary) {
        middle[middle.length - 1] = '...'
    }

    pages.splice(1, 0, ...middle)

    return pages
}

export const usePagination = ({
    currentPage,
    onChange,
    totalPages,
    maxDisplayed = totalPages,
}: UsePaginationParameters): PaginationItem[] => {
    const [activePage, setActivePage] = useControlledState({ value: currentPage, onChange })

    const goTo = useCallback((page: number) => () => setActivePage(page), [setActivePage])

    const pages: PaginationItem[] = getPages({ currentPage: activePage, totalPages, maxDisplayed }).map(page => {
        if (page === '...') {
            return {
                content: page,
                type: 'elipsis',
                active: false,
                disabled: true,
            }
        }

        return {
            content: page,
            type: 'page',
            active: currentPage === page,
            'aria-label': `Go to page ${page}`,
            disabled: false,
            onClick: goTo(page),
        }
    })

    return [
        {
            type: 'previous',
            active: false,
            content: 'Previous',
            'aria-label': 'Go to previous page',
            disabled: activePage === 1,
            onClick: goTo(activePage - 1),
        },
        ...pages,
        {
            type: 'next',
            active: false,
            content: 'Next',
            'aria-label': 'Go to next page',
            disabled: activePage === totalPages,
            onClick: goTo(activePage + 1),
        },
    ]
}
