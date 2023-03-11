import React from 'react'

import classNames from 'classnames'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import PageFirstIcon from 'mdi-react/PageFirstIcon'
import PageLastIcon from 'mdi-react/PageLastIcon'

import { Button } from '../Button'
import { Icon } from '../Icon'
import { Tooltip } from '../Tooltip'
import { Text } from '../Typography'

import styles from './PageSwitcher.module.scss'

export interface PageSwitcherProps {
    totalLabel?: string
    totalCount: null | number
    hasNextPage: null | boolean
    hasPreviousPage: null | boolean
    goToNextPage: () => Promise<void>
    goToPreviousPage: () => Promise<void>
    goToFirstPage: () => Promise<void>
    goToLastPage: () => Promise<void>
    onClick?: () => void
    className?: string
}

/**
 * PageSwitcher is used to render pagination control for a cursor-based
 * pagination.
 *
 * It works best together with the `usePageSwitcherPagination` hook and
 * is our recommended way of implementing pagination.
 */
export const PageSwitcher: React.FunctionComponent<React.PropsWithChildren<PageSwitcherProps>> = props => {
    const {
        className,
        totalLabel,
        totalCount,
        goToFirstPage,
        goToPreviousPage,
        goToNextPage,
        goToLastPage,
        hasPreviousPage,
        hasNextPage,
        onClick,
    } = props

    const [isLoadingPage, setIsLoadingPage] = React.useState(false)
    function withLoadingPage<T>(func: () => Promise<T>): () => Promise<void> {
        return async () => {
            try {
                onClick?.()
                setIsLoadingPage(true)
                await func()
            } finally {
                setIsLoadingPage(false)
            }
        }
    }

    const isPreviousPageDisabled = isLoadingPage || (hasPreviousPage !== null ? !hasPreviousPage : true)
    const isNextPageDisabled = isLoadingPage || (hasNextPage !== null ? !hasNextPage : true)

    if (isPreviousPageDisabled && isNextPageDisabled && !isLoadingPage) {
        return null
    }

    return (
        <nav className={className} aria-label="pagination">
            <ul className={styles.list}>
                <li>
                    <Tooltip content="First page">
                        <Button
                            aria-label="Go to first page"
                            className={classNames(styles.button, 'mx-3')}
                            variant="secondary"
                            outline={true}
                            disabled={isPreviousPageDisabled}
                            onClick={withLoadingPage(goToFirstPage)}
                        >
                            <Icon aria-hidden={true} as={PageFirstIcon} className={styles.firstPageButton} />
                        </Button>
                    </Tooltip>
                </li>
                <li>
                    <Button
                        className={classNames(styles.button, styles.previousButton, 'mx-1')}
                        aria-label="Go to previous page"
                        variant="secondary"
                        outline={true}
                        disabled={isPreviousPageDisabled}
                        onClick={withLoadingPage(goToPreviousPage)}
                    >
                        <Icon
                            aria-hidden={true}
                            as={ChevronLeftIcon}
                            className={classNames('mr-1', styles.previousButtonIcon)}
                        />
                        <span className={styles.previousButtonLabel}>Prev</span>
                    </Button>
                </li>
                <li>
                    <Button
                        className={classNames(styles.button, styles.nextButton, 'mx-1')}
                        aria-label="Go to next page"
                        variant="secondary"
                        outline={true}
                        disabled={isNextPageDisabled}
                        onClick={withLoadingPage(goToNextPage)}
                    >
                        <span className={styles.nextButtonLabel}>Next</span>
                        <Icon
                            aria-hidden={true}
                            as={ChevronRightIcon}
                            className={classNames('ml-1', styles.nextButtonIcon)}
                        />
                    </Button>
                </li>
                <li>
                    <Tooltip content="Last page">
                        <Button
                            aria-label="Go to last page"
                            className={classNames(styles.button, 'mx-3')}
                            variant="secondary"
                            outline={true}
                            disabled={isNextPageDisabled}
                            onClick={withLoadingPage(goToLastPage)}
                        >
                            <Icon aria-hidden={true} as={PageLastIcon} className={styles.lastPageButton} />
                        </Button>
                    </Tooltip>
                </li>
            </ul>
            {totalCount !== null && totalLabel !== undefined ? (
                <div className={styles.label}>
                    <Text className="text-muted mb-0" size="small">
                        Total{' '}
                        <Text weight="bold" as="strong">
                            {totalLabel}
                        </Text>
                        : {totalCount}
                    </Text>
                </div>
            ) : null}
        </nav>
    )
}
