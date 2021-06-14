import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'
import React, { useEffect, useState } from 'react'
import { Collapse } from 'reactstrap'

import { FilterLink, FilterLinkProps } from './FilterLink'
import styles from './SearchSidebarSection.module.scss'

export const SearchSidebarSection: React.FunctionComponent<{
    header: string
    children?: React.ReactElement[]
    className?: string
    showSearch?: boolean // Search only works if children are FilterLink
}> = ({ header, children = [], className, showSearch = false }) => {
    const [filter, setFilter] = useState('')

    // Clear filter when children change
    useEffect(() => setFilter(''), [children])

    const filteredChildren = children.filter(child => {
        if (child.type === FilterLink) {
            const props: FilterLinkProps = child.props as FilterLinkProps
            return (
                (props?.label).toLowerCase().includes(filter.toLowerCase()) ||
                (props?.value).toLowerCase().includes(filter.toLowerCase())
            )
        }
        return true
    })

    const [collapsed, setCollapsed] = useState(false)

    const searchVisible = showSearch && children.length > 1

    return children.length > 0 ? (
        <div className={classNames(styles.sidebarSection, className)}>
            <button
                type="button"
                className={classNames('btn btn-outline-secondary', styles.sidebarSectionCollapseButton)}
                onClick={() => setCollapsed(collapsed => !collapsed)}
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

                    <ul className={styles.sidebarSectionList}>
                        {filteredChildren.map((child, index) => (
                            <li key={child.key || index}>{child}</li>
                        ))}
                        {filteredChildren.length === 0 && (
                            <li className={classNames('text-muted', styles.sidebarSectionNoResults)}>No results</li>
                        )}
                    </ul>
                </div>
            </Collapse>
        </div>
    ) : null
}
