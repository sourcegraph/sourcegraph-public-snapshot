import React from 'react'
import classNames from 'classnames'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import useResizeObserver from 'use-resize-observer/polyfilled'
import { useDebounce } from '../../hooks/useDebounce'
import { omit } from 'lodash'
import { usePagination } from '../../hooks/usePagination'

interface PageButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    active?: boolean
}

const PageButton = React.forwardRef<HTMLButtonElement, PageButtonProps>(({ children, active, ...props }, reference) => (
    <button
        type="button"
        className={classNames('btn', 'mx-1', 'pagination__button', active && 'btn-primary')}
        aria-current={active}
        ref={reference}
        {...props}
    >
        {children}
    </button>
))

export interface PaginationProps {
    currentPage: number
    onPageChange: (page: number) => void
    totalPages: number
    className?: string
}

const validatePositiveInteger = (name: string, value: number): string[] => {
    if (value > 0) {
        return []
    }
    return [`${name} must have a value greater than 0`]
}

const validateProps = (props: PaginationProps): void => {
    const errors = [
        ...validatePositiveInteger('totalPages', props.totalPages),
        ...validatePositiveInteger('currentPage', props.currentPage),
    ]

    if (props.currentPage > props.totalPages) {
        errors.push('currentPage must be not be greater than totalPages')
    }

    if (errors.length > 0) {
        throw new Error(`Invalid configuration:\n${errors.map(error => `— ${error}`).join('\n')}`)
    }
}

export const Pagination: React.FunctionComponent<PaginationProps> = props => {
    validateProps(props)

    const { totalPages, currentPage, className, onPageChange } = props
    const { ref, width } = useDebounce(useResizeObserver(), 100)
    const shouldShrink = width !== undefined && width <= 576
    const pages = usePagination({
        currentPage,
        onChange: onPageChange,
        totalPages,
        maxDisplayed: shouldShrink ? 3 : 5,
    })

    return (
        <nav>
            <ul ref={ref} className={classNames('pagination', className)}>
                {pages.map((page, index) => {
                    if (page.type === 'previous' || page.type === 'next') {
                        return (
                            <li key={index}>
                                <PageButton {...omit(page, 'type')}>
                                    {page.type === 'previous' && (
                                        <ChevronLeftIcon className="icon-inline" aria-hidden={true} />
                                    )}
                                    <span className={classNames(shouldShrink && 'd-none')}>{page.content}</span>
                                    {page.type === 'next' && (
                                        <ChevronRightIcon className="icon-inline" aria-hidden={true} />
                                    )}
                                </PageButton>
                            </li>
                        )
                    }

                    return (
                        <li key={index}>
                            <PageButton {...omit(page, 'type')}>{page.content}</PageButton>
                        </li>
                    )
                })}
            </ul>
        </nav>
    )
}
