import classNames from 'classnames'
import React, { useEffect, useMemo, useState } from 'react'

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
                    const props = child.props as FilterLinkProps
                    return (
                        (props?.label).toLowerCase().includes(filter.toLowerCase()) ||
                        (props?.value).toLowerCase().includes(filter.toLowerCase())
                    )
                }
                return true
            }),
        [children, filter]
    )

    return children.length > 0 ? (
        <div className="mb-4">
            <h5 className="pb-2">{header}</h5>
            {showSearch && children.length > 1 && (
                <input
                    type="search"
                    placeholder="Find..."
                    value={filter}
                    onChange={event => setFilter(event.currentTarget.value)}
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
    ) : null
}
