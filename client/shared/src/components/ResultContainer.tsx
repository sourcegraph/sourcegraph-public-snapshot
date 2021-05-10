/* eslint jsx-a11y/click-events-have-key-events: warn, jsx-a11y/no-static-element-interactions: warn */
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import ChevronUpIcon from 'mdi-react/ChevronUpIcon'
import React, { useEffect, useState } from 'react'

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
     * This component does not accept children.
     */
    children?: never
}

const blockExpandAndCollapse = (event: React.MouseEvent<HTMLElement>): void => event.stopPropagation()

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
}) => {
    const [expanded, setExpanded] = useState(allExpanded || defaultExpanded)

    useEffect(() => setExpanded(allExpanded || defaultExpanded), [allExpanded, defaultExpanded])

    const toggle = (): void => setExpanded(expanded => !expanded)

    const Icon = icon
    return (
        <div className="test-search-result result-container" data-testid="result-container">
            {/* TODO: Fix accessibility issues.
                Issue: https://github.com/sourcegraph/sourcegraph/issues/19272 */}
            <div
                className={'result-container__header' + (collapsible ? ' result-container__header--collapsible' : '')}
                onClick={toggle}
            >
                <Icon className="icon-inline" />
                <div
                    className={`result-container__header-title ${titleClassName || ''}`}
                    data-testid="result-container-header"
                >
                    {collapsible ? (
                        // This is to ensure the onClick toggle handler doesn't get called
                        // We should be able to remove this if we refactor to seperate the toggle to its own button
                        <span onClick={blockExpandAndCollapse}>{title}</span>
                    ) : (
                        title
                    )}
                    {description && <span className="ml-2">{description}</span>}
                </div>
                {collapsible &&
                    (expanded ? (
                        <small className="result-container__toggle-matches-container">
                            {collapseLabel}
                            {collapseLabel && <ChevronUpIcon className="icon-inline" />}
                            {!collapseLabel && <ChevronDownIcon className="icon-inline" />}
                        </small>
                    ) : (
                        <small className="result-container__toggle-matches-container">
                            {expandLabel}
                            {expandLabel && <ChevronDownIcon className="icon-inline" />}
                            {!expandLabel && <ChevronRightIcon className="icon-inline" />}
                        </small>
                    ))}
            </div>
            {!expanded && collapsedChildren}
            {expanded && expandedChildren}
        </div>
    )
}
