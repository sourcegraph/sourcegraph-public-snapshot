import * as H from 'history'
import classNames from 'classnames'
import React, { useCallback, useEffect, useState } from 'react'
import { ButtonDropdown, DropdownMenu, DropdownToggle } from 'reactstrap'
import { CaseSensitivityProps, isContextFilterInQuery, PatternTypeProps, SearchContextProps } from '..'
import { SearchContextMenu } from './SearchContextMenu'
import { submitSearch } from '../helpers'
import { VersionContextProps } from '../../../../shared/src/search/util'

export interface SearchContextDropdownProps
    extends Omit<SearchContextProps, 'showSearchContext'>,
        Pick<PatternTypeProps, 'patternType'>,
        Pick<CaseSensitivityProps, 'caseSensitive'>,
        VersionContextProps {
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
                activation: undefined,
                searchParameters: [],
            })
        },
        [caseSensitive, history, query, patternType, versionContext]
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
            <ButtonDropdown isOpen={isOpen} toggle={toggleOpen}>
                <DropdownToggle
                    className={classNames('search-context-dropdown__button', 'dropdown-toggle', {
                        'search-context-dropdown__button--open': isOpen,
                    })}
                    color="link"
                    disabled={isDisabled}
                    data-tooltip={isDisabled ? 'Overridden by query' : ''}
                >
                    <code className="search-context-dropdown__button-content">
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
            </ButtonDropdown>
            <div className="search-context-dropdown__separator" />
        </>
    )
}
