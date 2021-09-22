import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'
import React, { useEffect, useState } from 'react'
import { Collapse } from 'reactstrap'

import { FilterLink, FilterLinkProps } from './FilterLink'
import styles from './SearchSidebarSection.module.scss'

export const SearchSidebarSection: React.FunctionComponent<{
    sectionId: string
    header: string
    children?: React.ReactElement | React.ReactElement[] | ((filter: string) => React.ReactElement)
    className?: string
    showSearch?: boolean // Search only works if children are FilterLink
    onToggle?: (id: string, open: boolean) => void
    startCollapsed?: boolean
    /**
     * Shown when the built-in search doesn't find any results.
     */
    noResultText?: React.ReactElement | string
    /**
     * Clear the search input whenever this value changes. This is supposed to
     * be used together with function children, which use the search input but
     * handle search on their own.
     * Defaults to the component's children.
     */
    clearSearchOnChange?: {}
}> = React.memo(
    ({
        sectionId,
        header,
        children = [],
        className,
        showSearch = false,
        onToggle,
        startCollapsed,
        noResultText = 'No results',
        clearSearchOnChange = children,
    }) => {
        const [filter, setFilter] = useState('')

        // Clears the filter whenever clearSearchOnChange changes (defaults to the
        // component's children)
        useEffect(() => setFilter(''), [clearSearchOnChange])

        let body
        let searchVisible = showSearch
        let visible = false

        if (typeof children === 'function') {
            visible = true
            body = children(filter)
        } else if (Array.isArray(children)) {
            visible = children.length > 0
            searchVisible = searchVisible && children.length > 1
            const childrenList = children as React.ReactElement[]

            const filteredChildren = searchVisible
                ? childrenList.filter(child => {
                      if (child.type === FilterLink) {
                          const props: FilterLinkProps = child.props as FilterLinkProps
                          return (
                              (props?.label).toLowerCase().includes(filter.toLowerCase()) ||
                              (props?.value).toLowerCase().includes(filter.toLowerCase())
                          )
                      }
                      return true
                  })
                : childrenList

            body = (
                <>
                    <ul className={styles.sidebarSectionList}>
                        {filteredChildren.map((child, index) => (
                            <li key={child.key || index}>{child}</li>
                        ))}
                        {filteredChildren.length === 0 && (
                            <li className={classNames('text-muted', styles.sidebarSectionNoResults)}>{noResultText}</li>
                        )}
                    </ul>
                </>
            )
        } else {
            visible = true
            body = children
        }

        const [collapsed, setCollapsed] = useState(startCollapsed)
        useEffect(() => setCollapsed(startCollapsed), [startCollapsed])

        return visible ? (
            <div className={classNames(styles.sidebarSection, className)}>
                <button
                    type="button"
                    className={classNames('btn btn-outline-secondary', styles.sidebarSectionCollapseButton)}
                    onClick={() =>
                        setCollapsed(collapsed => {
                            if (onToggle) {
                                onToggle(sectionId, !collapsed)
                            }
                            return !collapsed
                        })
                    }
                    aria-label={collapsed ? 'Expand' : 'Collapse'}
                >
                    <h5 className="flex-grow-1">{header}</h5>
                    {collapsed ? (
                        <ChevronLeftIcon className="icon-inline mr-1" />
                    ) : (
                        <ChevronDownIcon className="icon-inline mr-1" />
                    )}
                </button>

                <Collapse isOpen={!collapsed}>
                    <div className={classNames('pb-4', !searchVisible && 'border-top')}>
                        {searchVisible && (
                            <input
                                type="search"
                                placeholder="Find..."
                                aria-label="Find filters"
                                value={filter}
                                onChange={event => setFilter(event.currentTarget.value)}
                                data-testid="sidebar-section-search-box"
                                className={classNames('form-control form-control-sm', styles.sidebarSectionSearchBox)}
                            />
                        )}
                        {body}
                    </div>
                </Collapse>
            </div>
        ) : null
    }
)
