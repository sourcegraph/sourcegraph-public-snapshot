import ChevronDoubleLeftIcon from 'mdi-react/ChevronDoubleLeftIcon'
import ChevronDoubleRightIcon from 'mdi-react/ChevronDoubleRightIcon'
import React, { useCallback, useMemo, useState } from 'react'

interface Props {
    /**
     * Called to render the sidebar in either the expanded or collapsed state.
     */
    children: (expanded: boolean) => JSX.Element

    /**
     * If set, the expanded state is persisted in localStorage.
     */
    localStorageKey?: string

    /** The side of the screen where the sidebar is anchored. */
    side: 'left' | 'right'

    className?: string
    collapsedClassName?: string
    expandedClassName?: string
}

/**
 * A sidebar that can be collapsed.
 */
export const CollapsibleSidebar: React.FunctionComponent<Props> = ({
    localStorageKey,
    side,
    className,
    collapsedClassName,
    expandedClassName,
    children,
}) => {
    const initialIsExpanded = useMemo(
        () => (localStorageKey !== undefined ? localStorage.getItem(localStorageKey) !== null : true),
        [localStorageKey]
    )
    const [isExpanded, setIsExpanded] = useState(initialIsExpanded)
    const toggleIsExpanded = useCallback(() => {
        setIsExpanded(!isExpanded)
        if (localStorageKey !== undefined) {
            if (isExpanded) {
                localStorage.removeItem(localStorageKey)
            } else {
                localStorage.setItem(localStorageKey, 'expanded')
            }
        }
    }, [isExpanded, localStorageKey])

    const Icon =
        (isExpanded && side === 'left') || (!isExpanded && side === 'right')
            ? ChevronDoubleLeftIcon
            : ChevronDoubleRightIcon

    return (
        <div
            className={`collapsible-sidebar d-flex flex-column justify-content-between ${className} ${
                isExpanded ? expandedClassName : collapsedClassName
            }`}
        >
            <div>{children(isExpanded)}</div>
            <button
                className={`btn btn-link text-decoration-none py-3 border-top d-flex align-items-center ${
                    isExpanded ? 'px-2' : 'justify-content-center px-0'
                }`}
                title="Toggle sidebar"
                onClick={toggleIsExpanded}
            >
                <Icon className="icon-inline mr-1" />
                {isExpanded && <>&nbsp;Collapse sidebar</>}
            </button>
        </div>
    )
}
