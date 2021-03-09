import React from 'react'

interface PageButton extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    active?: boolean
}

const PageButton: React.FunctionComponent<PageButton> = ({ children, active, ...props }) => (
    <li>
        <button type="button" className={`btn mx-1 ${active ? 'btn-primary' : ''}`} {...props}>
            {children}
        </button>
    </li>
)

const range = (start: number, end: number) => {
    const length = end - start + 1
    return Array.from({ length }, (_, i) => start + i)
}

type Page = '...' | number

const getMiddlePages = (currentPage: number, maxPages: number): Page[] => {
    const maximumSiblingRange = 6
    const minimumBoundary = 2
    const maximumBoundary = maxPages - 1

    const siblingsStart = Math.max(
        Math.min(
            currentPage - 2,
            maxPages - maximumSiblingRange + 1 // Extend range when page is high
        ),
        minimumBoundary
    )

    const siblingsEnd = Math.min(
        Math.max(
            currentPage + 2,
            maximumSiblingRange // Extend range when page is low
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

const getPages = (currentPage: number, maxPages: number): Page[] => {
    const pages: Page[] = [1]

    if (maxPages > 1) {
        pages.push(maxPages)
    }

    if (maxPages > 2) {
        pages.splice(1, 0, ...getMiddlePages(currentPage, maxPages))
    }

    return pages
}

interface Props {
    currentPage: number
    onPageChange: (page: number) => void
    maxPages: number
}

export const PageSelector: React.FunctionComponent<Props> = ({ currentPage, onPageChange, maxPages }) => {
    const pages = getPages(currentPage, maxPages)

    return (
        <nav>
            <ul className="page-selector">
                <PageButton onClick={() => onPageChange(Math.max(currentPage - 1, 1))}>Previous</PageButton>
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
                <PageButton onClick={() => onPageChange(Math.min(currentPage + 1, maxPages))}>Next</PageButton>
            </ul>
        </nav>
    )
}
