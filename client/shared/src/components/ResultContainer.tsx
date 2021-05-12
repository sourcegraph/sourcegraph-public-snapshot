/* eslint jsx-a11y/click-events-have-key-events: warn, jsx-a11y/no-static-element-interactions: warn */
import ArrowCollapseUpIcon from 'mdi-react/ArrowCollapseUpIcon'
import ArrowExpandDownIcon from 'mdi-react/ArrowExpandDownIcon'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'
import ChevronUpIcon from 'mdi-react/ChevronUpIcon'
import React, { useEffect, useState } from 'react'

import { useRedesignToggle } from '../util/useRedesignToggle'

export interface Props {
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
}

/**
 * The container component for a result in the SearchResults component.
 */
export const ResultContainer: React.FunctionComponent<Props> = ({
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
}) => {
    const [isRedesignEnabled] = useRedesignToggle()
    const [expanded, setExpanded] = useState(allExpanded || defaultExpanded)

    useEffect(() => setExpanded(allExpanded || defaultExpanded), [allExpanded, defaultExpanded])

    const toggle = (): void => {
        if (collapsible) {
            setExpanded(expanded => !expanded)
        }
    }

    const Icon = icon
    return (
        <div className="test-search-result result-container" data-testid="result-container">
            <div className="result-container__header">
                <Icon className="icon-inline" />
                <div className="result-container__header-divider" />
                <div
                    className={`result-container__header-title ${titleClassName || ''}`}
                    data-testid="result-container-header"
                >
                    {title}
                    {description && <span className="ml-2">{description}</span>}
                </div>
                {matchCountLabel && isRedesignEnabled && (
                    <>
                        <small className="mr-1">{matchCountLabel}</small>
                        {collapsible && <div className="result-container__header-divider" />}
                    </>
                )}
                {collapsible &&
                    (expanded ? (
                        <button
                            type="button"
                            className="result-container__toggle-matches-container btn btn-sm btn-link"
                            onClick={toggle}
                        >
                            {collapseLabel && isRedesignEnabled && <ArrowCollapseUpIcon className="icon-inline mr-1" />}
                            {collapseLabel}
                            {collapseLabel && !isRedesignEnabled && <ChevronUpIcon className="icon-inline" />}
                            {!collapseLabel && <ChevronDownIcon className="icon-inline" />}
                        </button>
                    ) : (
                        <button
                            type="button"
                            className="result-container__toggle-matches-container btn btn-sm btn-link"
                            onClick={toggle}
                        >
                            {expandLabel && isRedesignEnabled && <ArrowExpandDownIcon className="icon-inline mr-1" />}
                            {expandLabel}
                            {expandLabel && !isRedesignEnabled && <ChevronDownIcon className="icon-inline" />}
                            {!expandLabel && <ChevronLeftIcon className="icon-inline" />}
                        </button>
                    ))}
            </div>
            {!expanded && collapsedChildren}
            {expanded && expandedChildren}
        </div>
    )
}
