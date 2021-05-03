import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'
import React, { useEffect, useMemo, useState } from 'react'
import { Collapse } from 'reactstrap'

import { FilterLink, FilterLinkProps } from './FilterLink'
import styles from './SearchSidebarSection.module.scss'

export const SearchSidebarSection: React.FunctionComponent<{
    header: string
    children?: React.ReactElement[]
    showSearch?: boolean // Search only works if children are FilterLink
}> = ({ header, children = [], showSearch = false }) => {
    const [filter, setFilter] = useState('')

    // Clear filter when children change
    useEffect(() => setFilter(''), [children])

    const filteredChildren = useMemo(
        () =>
            children.filter(child => {
                if (child.type === FilterLink) {
                    const props: FilterLinkProps = child.props as FilterLinkProps
                    return (
                        (props?.label).toLowerCase().includes(filter.toLowerCase()) ||
                        (props?.value).toLowerCase().includes(filter.toLowerCase())
                    )
                }
                return true
            }),
        [children, filter]
    )

    const [collapsed, setCollapsed] = useState(false)

    return children.length > 0 ? (
        <div>
            <button
                type="button"
                className={classNames('btn btn-outline-secondary', styles.sidebarSectionCollapseButton)}
                onClick={() => setCollapsed(collapsed => !collapsed)}
                aria-label={collapsed ? 'Expand' : 'Collapse'}
            >
                <h5 className="flex-grow-1">{header}</h5>
                {collapsed ? <ChevronLeftIcon className="icon-inline" /> : <ChevronDownIcon className="icon-inline" />}
            </button>

            <Collapse isOpen={!collapsed}>
                <div className="pb-4">
                    {showSearch && children.length > 1 && (
                        <input
                            type="search"
                            placeholder="Find..."
                            aria-label="Find filters"
                            value={filter}
                            onChange={event => setFilter(event.currentTarget.value)}
                            data-testid="sidebar-section-search-box"
                            className={classNames(
                                'form-control',
                                styles.sidebarSectionSearchBox,
                                'test-sidebar-section-search-box'
                            )}
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
