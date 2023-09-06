import React, { type ReactElement, useCallback } from 'react'

import classNames from 'classnames'

import {
    type BuildSearchQueryURLParameters,
    type QueryState,
    type SearchContextProps,
    createQueryExampleFromString,
    updateQueryWithFilterAndExample,
    EditorHint,
} from '@sourcegraph/shared/src/search'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { updateFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { containsLiteralOrPattern } from '@sourcegraph/shared/src/search/query/validate'
import type { SearchType } from '@sourcegraph/shared/src/search/stream'
import { Button, Link, createLinkUrl } from '@sourcegraph/wildcard'

import styles from './SearchFilterSection.module.scss'

export interface SearchTypeLinksProps extends Pick<SearchContextProps, 'selectedSearchContextSpec'> {
    query: string
    onNavbarQueryChange: (queryState: QueryState) => void
    buildSearchURLQueryFromQueryState: (queryParameters: BuildSearchQueryURLParameters) => string
    /**
     * Force search type links to be rendered as buttons.
     * Used e.g. in the VS Code extension to update search query state.
     */
    forceButton?: boolean
}

interface SearchTypeLinkProps extends SearchTypeLinksProps {
    type: SearchType
    children: string
    'data-testid'?: string
}

/**
 * SearchTypeLink renders to a Link which immediately triggers a new search when
 * clicked.
 */
const SearchTypeLink: React.FunctionComponent<React.PropsWithChildren<SearchTypeLinkProps>> = ({
    type,
    query,
    selectedSearchContextSpec,
    children,
    buildSearchURLQueryFromQueryState,
    'data-testid': dataTestID,
}) => {
    const builtURLQuery = buildSearchURLQueryFromQueryState({
        query: updateFilter(query, FilterType.type, type as string),
        searchContextSpec: selectedSearchContextSpec,
    })

    return (
        <Link
            to={createLinkUrl({ pathname: '/search', search: builtURLQuery })}
            className={styles.sidebarSectionListItem}
            data-testid={dataTestID}
        >
            <span className={styles.sidebarSectionListItemLabel}>{children}</span>
        </Link>
    )
}

interface SearchTypeButtonProps {
    children: string
    onClick: () => void
    'data-testid'?: string
}

/**
 * SearchTypeButton renders to a button which updates the query state without
 * triggering a search. This allows users to adjust the query.
 */
const SearchTypeButton: React.FunctionComponent<React.PropsWithChildren<SearchTypeButtonProps>> = ({
    children,
    onClick,
    'data-testid': dataTestID,
}) => (
    <Button
        className={classNames(styles.sidebarSectionListItem, styles.sidebarSectionButtonLink, 'flex-1')}
        value={children}
        onClick={onClick}
        variant="link"
        data-testid={dataTestID}
    >
        <span className={styles.sidebarSectionListItemLabel}>{children}</span>
    </Button>
)

/**
 * SearchSymbolButton either renders to a Link or a button, depending on whether
 * the search should be triggered immediately at click (if the query contains
 * patterns) or whether to allow the user to complete query and triggering it
 * themselves.
 */
const SearchSymbol: React.FunctionComponent<React.PropsWithChildren<Omit<SearchTypeLinkProps, 'type'>>> = props => {
    const type = 'symbol'
    const { query, onNavbarQueryChange } = props

    const setSymbolSearch = useCallback(() => {
        onNavbarQueryChange({
            query: updateFilter(query, FilterType.type, type),
        })
    }, [query, onNavbarQueryChange])

    if (!props.forceButton && containsLiteralOrPattern(query)) {
        return (
            <SearchTypeLink {...props} type={type}>
                {props.children}
            </SearchTypeLink>
        )
    }
    return <SearchTypeButton onClick={setSymbolSearch}>{props.children}</SearchTypeButton>
}

const repoExample = createQueryExampleFromString('{regexp-pattern}')

export const getSearchTypeLinks = (props: SearchTypeLinksProps): ReactElement[] => {
    function updateQueryWithRepoExample(): void {
        const updatedQuery = updateQueryWithFilterAndExample(props.query, FilterType.repo, repoExample, {
            singular: false,
            negate: false,
            emptyValue: true,
        })
        props.onNavbarQueryChange({
            query: updatedQuery.query,
            selectionRange: updatedQuery.placeholderRange,
            revealRange: updatedQuery.filterRange,
            hint: EditorHint.ShowSuggestions,
        })
    }

    const SearchTypeLinkOrButton = props.forceButton ? SearchTypeButton : SearchTypeLink

    /** Click handler for `SearchTypeLinkOrButton` (when rendered as button) */
    function updateQueryWithType(type: string): void {
        props.onNavbarQueryChange({
            query: updateFilter(props.query, FilterType.type, type),
        })
    }

    return [
        <SearchTypeButton onClick={updateQueryWithRepoExample} key="repo" data-testid="search-type-suggest">
            Search repos by org or name
        </SearchTypeButton>,
        <SearchSymbol {...props} key="symbol">
            Find a symbol
        </SearchSymbol>,
        <SearchTypeLinkOrButton {...props} type="diff" key="diff" onClick={() => updateQueryWithType('diff')}>
            Search diffs
        </SearchTypeLinkOrButton>,
        <SearchTypeLinkOrButton
            {...props}
            type="commit"
            key="commit"
            onClick={() => updateQueryWithType('commit')}
            data-testid="search-type-submit"
        >
            Search commit messages
        </SearchTypeLinkOrButton>,
    ]
}
