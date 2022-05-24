import React, { useCallback, useEffect, useState } from 'react'

import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'

import { Button, Collapse, CollapseHeader, CollapsePanel, Icon, Typography } from '@sourcegraph/wildcard'

import { FilterLink, FilterLinkProps } from './FilterLink'

import styles from './SearchSidebarSection.module.scss'

export const SearchSidebarSection: React.FunctionComponent<
    React.PropsWithChildren<{
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
    }>
> = React.memo(
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

        const [isOpened, setOpened] = useState(!startCollapsed)
        const handleOpenChange = useCallback(
            isOpen => {
                if (onToggle) {
                    onToggle(sectionId, isOpen)
                }

                setOpened(isOpen)
            },
            [onToggle, sectionId]
        )

        return visible ? (
            <div className={classNames(styles.sidebarSection, className)}>
                <Collapse isOpen={isOpened} onOpenChange={handleOpenChange}>
                    <CollapseHeader
                        as={Button}
                        className={styles.sidebarSectionCollapseButton}
                        aria-label={isOpened ? 'Collapse' : 'Expand'}
                        outline={true}
                        variant="secondary"
                    >
                        <Typography.H5 as={Typography.H2} className="flex-grow-1">
                            {header}
                        </Typography.H5>
                        <Icon
                            role="img"
                            aria-hidden={true}
                            className="mr-1"
                            as={isOpened ? ChevronDownIcon : ChevronLeftIcon}
                        />
                    </CollapseHeader>

                    <CollapsePanel>
                        <div className={classNames('pb-4', !searchVisible && 'border-top')}>
                            {searchVisible && (
                                <input
                                    type="search"
                                    placeholder="Find..."
                                    aria-label="Find filters"
                                    value={filter}
                                    onChange={event => setFilter(event.currentTarget.value)}
                                    data-testid="sidebar-section-search-box"
                                    className={classNames(
                                        'form-control form-control-sm',
                                        styles.sidebarSectionSearchBox
                                    )}
                                />
                            )}
                            {body}
                        </div>
                    </CollapsePanel>
                </Collapse>
            </div>
        ) : null
    }
)
