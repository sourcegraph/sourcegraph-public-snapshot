import * as H from 'history'
import classNames from 'classnames'
import React, { useCallback, useEffect, useState } from 'react'
import { CaseSensitivityProps, PatternTypeProps, SearchContextProps } from '..'
import { Dropdown, DropdownMenu, DropdownToggle } from 'reactstrap'
import { SearchContextMenu } from './SearchContextMenu'
import { SubmitSearchParameters } from '../helpers'
import { VersionContextProps } from '../../../../shared/src/search/util'
import { isContextFilterInQuery } from '../../../../shared/src/search/query/validate'

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
    } = props

    const [isOpen, setIsOpen] = useState(false)
    const toggleOpen = useCallback(() => setIsOpen(value => !value), [])
    const [isDisabled, setIsDisabled] = useState(false)

    // Disable the dropdown if the query contains a context filter
    useEffect(() => setIsDisabled(isContextFilterInQuery(query)), [query])

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
                        closeMenu={toggleOpen}
                    />
                </DropdownMenu>
            </Dropdown>
            <div className="search-context-dropdown__separator" />
        </>
    )
}
