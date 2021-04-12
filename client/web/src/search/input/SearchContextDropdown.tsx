import classNames from 'classnames'
import * as H from 'history'
import React, { useCallback, useEffect, useState } from 'react'
import { Dropdown, DropdownMenu, DropdownToggle } from 'reactstrap'

import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { filterExists } from '@sourcegraph/shared/src/search/query/validate'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'

import { CaseSensitivityProps, PatternTypeProps, SearchContextProps } from '..'
import { SubmitSearchParameters } from '../helpers'

import { SearchContextMenu } from './SearchContextMenu'

export interface SearchContextDropdownProps
    extends Omit<SearchContextProps, 'showSearchContext'>,
        Pick<PatternTypeProps, 'patternType'>,
        Pick<CaseSensitivityProps, 'caseSensitive'>,
        VersionContextProps {
    submitSearch: (args: SubmitSearchParameters) => void
    query: string
    history: H.History
}

export const SearchContextDropdown: React.FunctionComponent<SearchContextDropdownProps> = props => {
    const {
        history,
        patternType,
        caseSensitive,
        versionContext,
        query,
        selectedSearchContextSpec,
        setSelectedSearchContextSpec,
        submitSearch,
        fetchAutoDefinedSearchContexts,
        fetchSearchContexts,
    } = props

    const [isOpen, setIsOpen] = useState(false)
    const toggleOpen = useCallback(() => setIsOpen(value => !value), [])
    const [isDisabled, setIsDisabled] = useState(false)

    // Disable the dropdown if the query contains a context filter
    useEffect(() => setIsDisabled(filterExists(query, FilterType.context)), [query])

    const submitOnToggle = useCallback(
        (selectedSearchContextSpec: string): void => {
            if (query === '') {
                return
            }
            submitSearch({
                history,
                query,
                source: 'filter',
                patternType,
                caseSensitive,
                selectedSearchContextSpec,
                versionContext,
            })
        },
        [submitSearch, caseSensitive, history, query, patternType, versionContext]
    )

    const selectSearchContextSpec = useCallback(
        (spec: string): void => {
            submitOnToggle(spec)
            setSelectedSearchContextSpec(spec)
        },
        [submitOnToggle, setSelectedSearchContextSpec]
    )

    return (
        <>
            <Dropdown
                isOpen={isOpen}
                toggle={toggleOpen}
                a11y={false} /* Override default keyboard events in reactstrap */
            >
                <DropdownToggle
                    className={classNames(
                        'search-context-dropdown__button',
                        'dropdown-toggle',
                        'test-search-context-dropdown',
                        {
                            'search-context-dropdown__button--open': isOpen,
                        }
                    )}
                    color="link"
                    disabled={isDisabled}
                    data-tooltip={isDisabled ? 'Overridden by query' : ''}
                >
                    <code className="search-context-dropdown__button-content test-selected-search-context-spec">
                        <span className="search-filter-keyword">context:</span>
                        {selectedSearchContextSpec?.startsWith('@') ? (
                            <>
                                <span className="search-keyword">@</span>
                                {selectedSearchContextSpec?.slice(1)}
                            </>
                        ) : (
                            selectedSearchContextSpec
                        )}
                    </code>
                </DropdownToggle>
                <DropdownMenu>
                    <SearchContextMenu
                        {...props}
                        selectSearchContextSpec={selectSearchContextSpec}
                        fetchAutoDefinedSearchContexts={fetchAutoDefinedSearchContexts}
                        fetchSearchContexts={fetchSearchContexts}
                        closeMenu={toggleOpen}
                    />
                </DropdownMenu>
            </Dropdown>
            <div className="search-context-dropdown__separator" />
        </>
    )
}
