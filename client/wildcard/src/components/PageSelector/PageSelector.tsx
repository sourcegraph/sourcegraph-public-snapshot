import classNames from 'classnames'
import { omit } from 'lodash'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React from 'react'
import useResizeObserver from 'use-resize-observer'

import { createAggregateError } from '@sourcegraph/common'

import { useOffsetPagination, useDebounce } from '../../hooks'

import styles from './PageSelector.module.scss'

interface PageButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    active?: boolean
}

const PageButton: React.FunctionComponent<PageButtonProps> = ({ children, active, ...props }) => (
    <button
        type="button"
        className={classNames('btn', active ? 'btn-primary' : 'btn-link', 'mx-1', styles.button)}
        aria-current={active}
        {...props}
    >
        {children}
    </button>
)

const isMobileViewport = (width: number): boolean => width < 576

export interface PageSelectorProps {
    /** Current active page */
    currentPage: number
    /** Fired on page change */
    onPageChange: (page: number) => void
    /** Maximum pages to use */
    totalPages: number
    className?: string
}

const validatePositiveInteger = (name: string, value: number): Error[] => {
    if (value > 0) {
        return []
    }
    return [new Error(`${name} must have a value greater than 0`)]
}

const validateProps = (props: PageSelectorProps): void => {
    const errors = [
        ...validatePositiveInteger('totalPages', props.totalPages),
        ...validatePositiveInteger('currentPage', props.currentPage),
    ]

    if (props.currentPage > props.totalPages) {
        errors.push(new Error('currentPage must be not be greater than totalPages'))
    }

    if (errors.length > 0) {
        throw createAggregateError(errors)
    }
}

/**
 * PageSelector should be used to render offset-pagination controls.
 * It is a controlled-component, the `currentPage` should be controlled by the consumer.
 */
export const PageSelector: React.FunctionComponent<PageSelectorProps> = props => {
    validateProps(props)

    const { totalPages, currentPage, className, onPageChange } = props
    const { ref, width } = useDebounce(useResizeObserver(), 100)
    const shouldShrink = width !== undefined && isMobileViewport(width)
    const pages = useOffsetPagination({
        page: currentPage,
        onChange: onPageChange,
        totalPages,
        maxDisplayed: shouldShrink ? 3 : 5,
    })

    return (
        <nav>
            <ul ref={ref} className={classNames(styles.list, className)}>
                {pages.map((page, index) => {
                    const key = page.type === 'page' ? page.content : `${page.type}${index}`
                    if (page.type === 'previous' || page.type === 'next') {
                        return (
                            <li key={key}>
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
                        <li key={key}>
                            <PageButton {...omit(page, 'type')}>{page.content}</PageButton>
                        </li>
                    )
                })}
            </ul>
        </nav>
    )
}
