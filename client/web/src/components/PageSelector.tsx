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

const getPages = (currentPage: number, maxPages: number): Page[] => {
    const maxSiblingRange = 5
    const boundary = 3
    const upperBoundary = maxPages - boundary + 1

    const siblingsStart = Math.max(
        Math.min(
            currentPage - 1,
            maxPages - maxSiblingRange + 1 // Extend range when page is high
        ),
        boundary // Minimum boundary
    )

    const siblingsEnd = Math.min(
        Math.max(
            currentPage + 1,
            maxSiblingRange // Extend range when page is low
        ),
        upperBoundary // Maximum boundary
    )

    const middle: Page[] = [
        siblingsStart === boundary ? boundary - 1 : '...',
        ...range(siblingsStart, siblingsEnd),
        siblingsEnd === upperBoundary ? upperBoundary + 1 : '...',
    ]

    return [1, ...middle, maxPages]
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
                <PageButton onClick={() => onPageChange(currentPage - 1)}>Previous</PageButton>
                {pages.map((page, i) => (
                    <PageButton key={i} disabled={page === '...'} onClick={() => page !== '...' && onPageChange(page)}>
                        {page}
                    </PageButton>
                ))}
                <PageButton onClick={() => onPageChange(currentPage + 1)}>Next</PageButton>
            </ul>
        </nav>
    )
}
