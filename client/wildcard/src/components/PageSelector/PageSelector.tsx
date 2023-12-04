import React from 'react'

import { mdiChevronLeft, mdiChevronRight } from '@mdi/js'
import classNames from 'classnames'
import { omit } from 'lodash'
import useResizeObserver from 'use-resize-observer'

import { createAggregateError } from '@sourcegraph/common'

import { useOffsetPagination, useDebounce } from '../../hooks'
import { Button } from '../Button'
import { Icon } from '../Icon'

import styles from './PageSelector.module.scss'

interface PageButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    active?: boolean
}

const PageButton: React.FunctionComponent<React.PropsWithChildren<PageButtonProps>> = ({
    children,
    active,
    ...props
}) => (
    <Button
        className={classNames('mx-1', styles.button)}
        variant={active ? 'primary' : 'link'}
        aria-current={active}
        {...props}
    >
        {children}
    </Button>
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
export const PageSelector: React.FunctionComponent<React.PropsWithChildren<PageSelectorProps>> = props => {
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
                                    {page.type === 'previous' && <Icon aria-hidden={true} svgPath={mdiChevronLeft} />}
                                    <span className={classNames(shouldShrink && 'd-none')}>{page.content}</span>
                                    {page.type === 'next' && <Icon aria-hidden={true} svgPath={mdiChevronRight} />}
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
