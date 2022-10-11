import React from 'react'

import classNames from 'classnames'

import { ForwardReferenceExoticComponent } from '@sourcegraph/wildcard'

import { formatRepositoryStarCount } from '../util/stars'

import { CodeHostIcon } from './CodeHostIcon'
import { SearchResultStar } from './SearchResultStar'

import styles from './ResultContainer.module.scss'

export interface ResultContainerProps {
    index: number
    title: React.ReactNode
    titleClassName?: string
    resultClassName?: string
    repoStars?: number
    resultType?: string
    repoName: string
    className?: string
    onResultClicked?: () => void
}

/**
 * The container component for a result in the SearchResults component.
 */
export const ResultContainer: ForwardReferenceExoticComponent<
    React.ElementType,
    React.PropsWithChildren<ResultContainerProps>
> = React.forwardRef(function ResultContainer(props, reference) {
    const {
        children,
        title,
        titleClassName,
        resultClassName,
        index,
        repoStars,
        resultType,
        repoName,
        className,
        as: Component = 'div',
        onResultClicked,
    } = props

    const formattedRepositoryStarCount = formatRepositoryStarCount(repoStars)

    const trackReferencePanelClick = (): void => onResultClicked?.()

    return (
        <Component
            className={classNames('test-search-result', styles.resultContainer, className)}
            data-testid="result-container"
            data-result-type={resultType}
            onClick={trackReferencePanelClick}
            ref={reference}
        >
            <article aria-labelledby={`result-container-${index}`}>
                <div className={styles.header} id={`result-container-${index}`}>
                    <CodeHostIcon repoName={repoName} className="text-muted flex-shrink-0 mr-1" />
                    <div
                        className={classNames(styles.headerTitle, titleClassName)}
                        data-testid="result-container-header"
                    >
                        {title}
                    </div>
                    {formattedRepositoryStarCount && (
                        <span className="d-flex align-items-center">
                            <SearchResultStar aria-label={`${repoStars} stars`} />
                            <span aria-hidden={true}>{formattedRepositoryStarCount}</span>
                        </span>
                    )}
                </div>
                <div className={classNames(styles.result, resultClassName)}>{children}</div>
            </article>
        </Component>
    )
})
