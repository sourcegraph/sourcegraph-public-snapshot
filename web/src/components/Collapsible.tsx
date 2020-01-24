import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import React, { useCallback, useState } from 'react'
import classNames from 'classnames'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'

interface Props {
    /**
     * Content in the always-visible, single-line title bar.
     */
    title: React.ReactNode

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

    /**
     * Whether the whole title section should be clickable to expand the content
     */
    wholeTitleClickable?: boolean
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
    wholeTitleClickable = true,
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
                <button
                    type="button"
                    className={classNames(
                        'btn btn-icon collapsible__expand-btn',
                        wholeTitleClickable && 'stretched-link'
                    )}
                    aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                    onClick={toggleIsExpanded}
                >
                    {isExpanded ? (
                        <ChevronDownIcon className="icon-inline" aria-label="Close section" />
                    ) : (
                        <ChevronRightIcon className="icon-inline" aria-label="Expand section" />
                    )}
                </button>
                <span className={titleClassName}>{title}</span>
            </div>
            {isExpanded && children}
        </div>
    )
}
