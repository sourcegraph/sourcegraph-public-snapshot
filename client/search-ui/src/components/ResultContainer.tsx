/* eslint jsx-a11y/click-events-have-key-events: warn, jsx-a11y/no-static-element-interactions: warn */
import React, { useEffect, useState } from 'react'

import classNames from 'classnames'
import ArrowCollapseUpIcon from 'mdi-react/ArrowCollapseUpIcon'
import ArrowExpandDownIcon from 'mdi-react/ArrowExpandDownIcon'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'

import { Button, Icon } from '@sourcegraph/wildcard'

import { formatRepositoryStarCount } from '../util/stars'

import { CodeHostIcon } from './CodeHostIcon'
import { SearchResultStar } from './SearchResultStar'

import styles from './ResultContainer.module.scss'

export interface ResultContainerProps {
    /**
     * Whether the result container's children are visible by default.
     * The header is always visible even when the component is not expanded.
     */
    defaultExpanded?: boolean

    /** Expand all results */
    allExpanded?: boolean

    /**
     * Whether the result container can be collapsed. If false, its children
     * are always displayed, and no expand/collapse actions are shown.
     */
    collapsible?: boolean

    /**
     * The icon to show left to the title.
     */
    icon: React.ComponentType<{ className?: string }>

    /**
     * The title component.
     */
    title: React.ReactFragment

    /**
     * CSS class name to apply to the title element.
     */
    titleClassName?: string

    /** The content to display next to the title. */
    description?: React.ReactFragment

    /**
     * The content of the result displayed underneath the result container's
     * header when collapsed.
     */
    collapsedChildren?: React.ReactFragment

    /**
     * The content of the result displayed underneath the result container's
     * header when expanded.
     */
    expandedChildren?: React.ReactFragment

    /**
     * The label to display next to the collapse button
     */
    collapseLabel?: string

    /**
     * The label to display next to the expand button
     */
    expandLabel?: string

    /**
     * The total number of matches to display
     */
    matchCountLabel?: string

    /**
     * This component does not accept children.
     */
    children?: never

    /**
     * The result type
     */
    resultType?: string

    /**
     * The name of the repository
     */
    repoName: string

    /**
     * The number of stars for the result's associated repo
     */
    repoStars?: number

    /**
     * The time the repo was last updated from the code host
     */
    repoLastFetched?: string

    /**
     * Click event for when the result is clicked
     */
    onResultClicked?: () => void

    /**
     * CSS class name to be applied to the component
     */
    className?: string
}

/**
 * The container component for a result in the SearchResults component.
 */
export const ResultContainer: React.FunctionComponent<React.PropsWithChildren<ResultContainerProps>> = ({
    defaultExpanded,
    allExpanded,
    collapsible,
    collapseLabel,
    expandLabel,
    collapsedChildren,
    expandedChildren,
    icon,
    title,
    titleClassName,
    description,
    matchCountLabel,
    repoName,
    repoStars,
    onResultClicked,
    className,
    resultType,
}) => {
    const [expanded, setExpanded] = useState(allExpanded || defaultExpanded)
    const formattedRepositoryStarCount = formatRepositoryStarCount(repoStars)

    useEffect(() => setExpanded(allExpanded || defaultExpanded), [allExpanded, defaultExpanded])

    const toggle = (): void => {
        if (collapsible) {
            setExpanded(expanded => !expanded)
        }
    }

    const trackReferencePanelClick = (): void => {
        if (onResultClicked) {
            onResultClicked()
        }
    }
    return (
        <div
            className={classNames('test-search-result', styles.resultContainer, className)}
            data-testid="result-container"
            data-result-type={resultType}
            data-expanded={allExpanded}
            onClick={trackReferencePanelClick}
            role="none"
        >
            <div className={styles.header}>
                <Icon
                    role="img"
                    className="flex-shrink-0"
                    as={icon}
                    aria-label={resultType ? `${resultType} result` : undefined}
                />
                <div className={classNames('mx-1', styles.headerDivider)} />
                <CodeHostIcon repoName={repoName} className="text-muted flex-shrink-0" />
                <div className={classNames(styles.headerTitle, titleClassName)} data-testid="result-container-header">
                    {title}
                    {description && <span className={classNames('ml-2', styles.headerDescription)}>{description}</span>}
                </div>
                {matchCountLabel && (
                    <>
                        <small>{matchCountLabel}</small>
                        {collapsible && <div className={classNames('mx-2', styles.headerDivider)} />}
                    </>
                )}
                {collapsible && (
                    <Button
                        data-testid="toggle-matches-container"
                        className={classNames('py-0', styles.toggleMatchesContainer)}
                        onClick={toggle}
                        variant="link"
                        size="sm"
                    >
                        {expanded ? (
                            <>
                                {collapseLabel && (
                                    <Icon role="img" className="mr-1" as={ArrowCollapseUpIcon} aria-hidden={true} />
                                )}
                                {collapseLabel}
                                {!collapseLabel && <Icon role="img" as={ChevronDownIcon} aria-hidden={true} />}
                            </>
                        ) : (
                            <>
                                {expandLabel && (
                                    <Icon role="img" className="mr-1" as={ArrowExpandDownIcon} aria-hidden={true} />
                                )}
                                {expandLabel}
                                {!expandLabel && <Icon role="img" as={ChevronLeftIcon} aria-hidden={true} />}
                            </>
                        )}
                    </Button>
                )}
                {matchCountLabel && formattedRepositoryStarCount && (
                    <div className={classNames('mx-2', styles.headerDivider)} />
                )}
                {formattedRepositoryStarCount && (
                    <>
                        <SearchResultStar aria-label={`${repoStars} stars`} />
                        <span aria-hidden={true}>{formattedRepositoryStarCount}</span>
                    </>
                )}
            </div>
            {!expanded && collapsedChildren}
            {expanded && expandedChildren}
        </div>
    )
}
