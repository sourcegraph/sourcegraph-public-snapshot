import React, { useCallback, useEffect, useState, memo, type FC, type ReactElement, type ReactNode } from 'react'

import { mdiChevronDown, mdiChevronUp } from '@mdi/js'
import classNames from 'classnames'

import { Button, Collapse, CollapseHeader, CollapsePanel, Icon, H5, Input, H3 } from '@sourcegraph/wildcard'

import { FilterLink, type FilterLinkProps } from './FilterLink'

import styles from './SearchFilterSection.module.scss'

export { styles }

export interface SearchFilterSectionProps {
    sectionId: string
    header: string
    postHeader?: ReactNode // Additional content to display after the header
    children?: React.ReactNode | React.ReactNode[] | ((filter: string) => React.ReactNode)
    className?: string

    onToggle?: (id: string, open: boolean) => void
    startCollapsed?: boolean

    forcedRender?: boolean

    /**
     * Minimal number of items to render the filter section. Defaults to 1.
     * This prop is set to 2 for repositories and languages filter sections:
     * when we have only one repo or language, the section shouldn't be rendered.
     */
    minItems?: number

    /**
     * If present, search is shown.
     * Search only works if children are FilterLink.
     */
    searchOptions?: {
        /**
         * Accessible label for the search box. These should all be unique.
         */
        ariaLabel: string

        /**
         * Shown when the built-in search doesn't find any results.
         */
        noResultText?: (links: ReactElement[]) => ReactNode

        /**
         * Clear the search input whenever this value changes. This is supposed to
         * be used together with function children, which use the search input but
         * handle search on their own.
         * Defaults to the component's children.
         */
        clearSearchOnChange?: unknown
    }
}

const defaultNoResult = (): string => 'No results'

/**
 * A wrapper UI component for rendering list of filters links (FilterLink) or any other custom
 * UI filter components. It may add search box for runtime filtering child items.
 *
 * Note: It's an internal component and used only in SearchSidebarSection component.
 *
 * TODO: Refactor this component see https://github.com/sourcegraph/sourcegraph/issues/40481
 */
export const SearchFilterSection: FC<SearchFilterSectionProps> = memo(props => {
    const {
        sectionId,
        header,
        postHeader,
        children = [],
        className,
        searchOptions,
        forcedRender = true,
        onToggle,
        startCollapsed,
        minItems = 1,
    } = props

    const { ariaLabel = '', noResultText = defaultNoResult, clearSearchOnChange = children } = searchOptions ?? {}

    const [filter, setFilter] = useState('')

    // Clears the filter whenever clearSearchOnChange changes (defaults to the
    // component's children)
    useEffect(() => setFilter(''), [clearSearchOnChange])

    let body
    let searchVisible = !!searchOptions
    let visible = false

    // Supports render props approach
    if (typeof children === 'function') {
        visible = true
        body = children(filter)

        // Supports list-like children, it's used when we need to render just a flat list of
        // items (usually it's FilterLink components)
    } else if (Array.isArray(children)) {
        // Sometimes we don't need to render filter section with just one item (example - repositories filter section)
        visible = children.length >= minItems

        // We don't need to have a search UI if we're dealing with only one item
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
                        <li className={classNames('text-muted', styles.sidebarSectionNoResults)}>
                            {noResultText(childrenList)}
                        </li>
                    )}
                </ul>
            </>
        )
    } else {
        // Just render children as it is without any checks or additional UI if child is regular React component
        visible = true
        body = children
    }

    const [isOpened, setOpened] = useState(!startCollapsed)
    const handleOpenChange = useCallback(
        (isOpen: boolean) => {
            if (onToggle) {
                onToggle(sectionId, isOpen)
            }

            setOpened(isOpen)
        },
        [onToggle, sectionId]
    )

    return visible ? (
        <article
            aria-labelledby={`search-sidebar-section-header-${sectionId}`}
            className={classNames(styles.sidebarSection, className)}
        >
            <Collapse isOpen={isOpened} onOpenChange={handleOpenChange}>
                <CollapseHeader
                    as={Button}
                    className={styles.sidebarSectionCollapseButton}
                    aria-label={`${isOpened ? 'Collapse' : 'Expand'} ${header}`}
                    outline={true}
                    variant="secondary"
                >
                    <header className={styles.sidebarSectionHeader}>
                        <H5 as={H3} id={`search-sidebar-section-header-${sectionId}`}>
                            {header}
                        </H5>
                        <H5 as="span">{postHeader}</H5>
                    </header>
                    <Icon aria-hidden={true} svgPath={isOpened ? mdiChevronUp : mdiChevronDown} />
                </CollapseHeader>

                <CollapsePanel forcedRender={forcedRender}>
                    <div className="pb-4">
                        {searchVisible && (
                            <Input
                                type="search"
                                placeholder="Find..."
                                aria-label={ariaLabel}
                                value={filter}
                                onChange={event => setFilter(event.currentTarget.value)}
                                data-testid="sidebar-section-search-box"
                                inputClassName={styles.sidebarSectionSearchBox}
                                variant="small"
                            />
                        )}
                        {body}
                    </div>
                </CollapsePanel>
            </Collapse>
        </article>
    ) : null
})

SearchFilterSection.displayName = 'SearchSidebarSection'
