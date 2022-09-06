/* eslint jsx-a11y/click-events-have-key-events: warn, jsx-a11y/no-static-element-interactions: warn */
import React, { useCallback, useEffect, useRef, useState } from 'react'

import { mdiArrowCollapseUp, mdiChevronDown, mdiArrowExpandDown, mdiChevronLeft, mdiChevronUp } from '@mdi/js'
import classNames from 'classnames'

import { useCoreWorkflowImprovementsEnabled } from '@sourcegraph/shared/src/settings/useCoreWorkflowImprovementsEnabled'
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
    title: React.ReactNode

    /**
     * CSS class name to apply to the title element.
     */
    titleClassName?: string

    /** The content to display next to the title. */
    description?: React.ReactNode

    /**
     * The content of the result displayed underneath the result container's
     * header when collapsed.
     */
    collapsedChildren?: React.ReactNode
    /**
     * The content of the result displayed underneath the result container's
     * header when expanded.
     */
    expandedChildren?: React.ReactNode
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

    resultsClassName?: string

    as?: React.ElementType
    index: number
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
    resultsClassName,
    resultType,
    as: Component = 'div',
    index,
}) => {
    const [coreWorkflowImprovementsEnabled] = useCoreWorkflowImprovementsEnabled()
    const [expanded, setExpanded] = useState(allExpanded || defaultExpanded)
    const formattedRepositoryStarCount = formatRepositoryStarCount(repoStars)

    useEffect(() => setExpanded(allExpanded || defaultExpanded), [allExpanded, defaultExpanded])

    const rootRef = useRef<HTMLElement>(null)
    const toggle = useCallback((): void => {
        if (collapsible) {
            setExpanded(expanded => !expanded)
        }

        // Scroll back to top of result when collapsing
        if (coreWorkflowImprovementsEnabled && expanded) {
            setTimeout(() => {
                const reducedMotion = !window.matchMedia('(prefers-reduced-motion: no-preference)').matches
                rootRef.current?.scrollIntoView({ block: 'nearest', behavior: reducedMotion ? 'auto' : 'smooth' })
            }, 0)
        }
    }, [collapsible, coreWorkflowImprovementsEnabled, expanded])

    const trackReferencePanelClick = (): void => {
        if (onResultClicked) {
            onResultClicked()
        }
    }
    return (
        <Component
            className={classNames('test-search-result', styles.resultContainer, className)}
            data-testid="result-container"
            data-result-type={resultType}
            data-expanded={allExpanded}
            data-collapsible={collapsible}
            onClick={trackReferencePanelClick}
            ref={rootRef}
        >
            <article aria-labelledby={`result-container-${index}`}>
                <div className={styles.header} id={`result-container-${index}`}>
                    {!coreWorkflowImprovementsEnabled && (
                        <>
                            <Icon
                                className="flex-shrink-0"
                                as={icon}
                                {...(resultType
                                    ? {
                                          'aria-label': `${resultType} result`,
                                      }
                                    : {
                                          'aria-hidden': true,
                                      })}
                            />
                            <div className={classNames('mx-1', styles.headerDivider)} />
                        </>
                    )}
                    <CodeHostIcon repoName={repoName} className="text-muted flex-shrink-0 mr-1" />
                    <div
                        className={classNames(styles.headerTitle, titleClassName)}
                        data-testid="result-container-header"
                    >
                        {title}
                        {description && (
                            <span className={classNames('ml-2', styles.headerDescription)}>{description}</span>
                        )}
                    </div>
                    {!coreWorkflowImprovementsEnabled && matchCountLabel && (
                        <span className="d-flex align-items-center">
                            <small>{matchCountLabel}</small>
                            {collapsible && <div className={classNames('mx-2', styles.headerDivider)} />}
                        </span>
                    )}
                    {!coreWorkflowImprovementsEnabled && collapsible && (
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
                                        <Icon className="mr-1" aria-hidden={true} svgPath={mdiArrowCollapseUp} />
                                    )}
                                    {collapseLabel}
                                    {!collapseLabel && <Icon aria-hidden={true} svgPath={mdiChevronDown} />}
                                </>
                            ) : (
                                <>
                                    {expandLabel && (
                                        <Icon className="mr-1" aria-hidden={true} svgPath={mdiArrowExpandDown} />
                                    )}
                                    {expandLabel}
                                    {!expandLabel && <Icon aria-hidden={true} svgPath={mdiChevronLeft} />}
                                </>
                            )}
                        </Button>
                    )}
                    {!coreWorkflowImprovementsEnabled && matchCountLabel && formattedRepositoryStarCount && (
                        <div className={classNames('mx-2', styles.headerDivider)} />
                    )}
                    {formattedRepositoryStarCount && (
                        <span className="d-flex align-items-center">
                            <SearchResultStar aria-label={`${repoStars} stars`} />
                            <span aria-hidden={true}>{formattedRepositoryStarCount}</span>
                        </span>
                    )}
                </div>
                <div
                    className={classNames(
                        coreWorkflowImprovementsEnabled && styles.collapsibleResults,
                        resultsClassName
                    )}
                >
                    <div>{expanded ? expandedChildren : collapsedChildren}</div>
                    {coreWorkflowImprovementsEnabled && collapsible && (
                        <button
                            type="button"
                            className={classNames(
                                styles.toggleMatchesButton,
                                expanded && styles.toggleMatchesButtonExpanded
                            )}
                            onClick={toggle}
                            data-testid="toggle-matches-container"
                        >
                            <Icon aria-hidden={true} svgPath={expanded ? mdiChevronUp : mdiChevronDown} />
                            <span className={styles.toggleMatchesButtonText}>
                                {expanded ? collapseLabel : expandLabel}
                            </span>
                        </button>
                    )}
                </div>
            </article>
        </Component>
    )
}
