import { useCallback } from 'react'

import { range } from 'lodash'

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

interface GetPagesParameters {
    activePage: number
    totalPages: number
    maxDisplayed: number
}

const getPages = ({ activePage, totalPages, maxDisplayed }: GetPagesParameters): Page[] => {
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
        values: [activePage - minDisplayed / 2, totalPages - maxDisplayed],
        minimum: minimumBoundary,
    })
    const siblingsEnd = highest({
        values: [activePage + minDisplayed / 2, maxDisplayed + 1],
        maximum: maximumBoundary,
    })

    const middle: Page[] = range(siblingsStart, siblingsEnd + 1)

    if (middle[0] !== minimumBoundary) {
        middle[0] = '...'
    }

    if (middle.at(-1) !== maximumBoundary) {
        middle[middle.length - 1] = '...'
    }

    pages.splice(1, 0, ...middle)

    return pages
}

interface UseOffsetPaginationParameters {
    page: number
    onChange?: (page: number) => void
    totalPages: number
    maxDisplayed: number
}

/**
 * useOffsetPagination is a hook to easily manage offset-pagination logic.
 * This hook is capable of controlling its own state, however it is possible to override this
 * by listening to the component state with `onChange` and updating `page` manually
 */
export function useOffsetPagination({
    page,
    onChange,
    totalPages,
    maxDisplayed = totalPages,
}: UseOffsetPaginationParameters): PaginationItem[] {
    const [activePage, setActivePage] = useControlledState({
        value: page,
        onChange,
    })

    const goTo = useCallback((page: number) => () => setActivePage(page), [setActivePage])

    const pages: PaginationItem[] = getPages({ activePage, totalPages, maxDisplayed }).map(page => {
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
            active: activePage === page,
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
