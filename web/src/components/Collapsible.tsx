import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronUpIcon from 'mdi-react/ChevronUpIcon'
import React, { useCallback, useState } from 'react'

interface Props {
    /**
     * Content in the always-visible, single-line title bar.
     */
    title: string

    /**
     * Optional children that appear below the title bar that can be expanded/collapsed. If present,
     * a button that expands or collapses the children will be shown.
     */
    children?: React.ReactFragment

    /**
     * Whether the children are expanded and visible by default.
     */
    defaultExpanded?: boolean

    className?: string
    titleClassName?: string
}

/**
 * Collapsible is an element with a title that is always displayed and children that are displayed
 * only when expanded.
 */
export const Collapsible: React.FunctionComponent<Props> = ({
    title,
    children,
    defaultExpanded = false,
    className = '',
    titleClassName = '',
}) => {
    const [isExpanded, setIsExpanded] = useState(defaultExpanded)
    const toggleIsExpanded = useCallback<React.MouseEventHandler<HTMLButtonElement>>(
        e => {
            e.preventDefault()
            setIsExpanded(!isExpanded)
        },
        [isExpanded]
    )

    return (
        <div className={className}>
            <div
                className={`d-flex justify-content-between align-items-center position-relative ${
                    isExpanded ? 'mb-3' : ''
                }`}
            >
                <span className={titleClassName}>{title}</span>
                <button
                    type="button"
                    className="btn btn-icon stretched-link"
                    aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                    onClick={toggleIsExpanded}
                >
                    {isExpanded ? (
                        <ChevronUpIcon className="icon-inline" aria-label="Close section" />
                    ) : (
                        <ChevronDownIcon className="icon-inline" aria-label="Expand section" />
                    )}
                </button>
            </div>
            {isExpanded && children}
        </div>
    )
}
