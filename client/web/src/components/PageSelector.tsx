import React from 'react'

interface ButtonProps {
    active?: boolean
    onClick?: () => void
    disabled?: boolean
    className?: string
}

const PageSelectorButton: React.FunctionComponent<ButtonProps> = ({ children, disabled, onClick, active }) => (
    <li>
        <button
            type="button"
            onClick={onClick}
            className={`btn mx-1 ${active ? 'btn-primary' : ''}`}
            disabled={disabled}
        >
            {children}
        </button>
    </li>
)

interface Props {
    currentPage: number
    onPageChange: (page: number) => void
    maxPages: number
}

const range = (start: number, end: number) => {
    const length = end - start + 1
    return Array.from({ length }, (_, i) => start + i)
}

type Page = '...' | number

const getPages = (currentPage: number, maxPages: number): Page[] => {
    const siblingCount = 1
    const boundary = 3
    const upperBoundary = maxPages - boundary + 1

    const siblingsStart = Math.max(
        Math.min(
            currentPage - siblingCount,
            upperBoundary - siblingCount * 2 // Lower boundary when page is high
        ),
        boundary // Minimum boundary
    )

    const siblingsEnd = Math.min(
        Math.max(
            currentPage + siblingCount,
            siblingCount * 2 + boundary // Upper boundary when page is low
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

export const PageSelector: React.FunctionComponent<Props> = ({ currentPage, onPageChange, maxPages }) => {
    const pages = getPages(currentPage, maxPages)

    return (
        <nav>
            <ul className="page-selector">
                <PageSelectorButton onClick={() => onPageChange(currentPage - 1)}>Previous</PageSelectorButton>
                {pages.map((page, i) => (
                    <li key={i}>
                        <button
                            type="button"
                            disabled={page === '...'}
                            onClick={() => page !== '...' && onPageChange(page)}
                            className={`btn mx-1 ${page === currentPage ? 'btn-primary' : ''}`}
                        >
                            {page}
                        </button>
                    </li>
                ))}
                <PageSelectorButton onClick={() => onPageChange(currentPage + 1)}>Next</PageSelectorButton>
            </ul>
        </nav>
    )
}
