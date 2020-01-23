import H from 'history'
import React, { useCallback } from 'react'
import { NavLink } from 'react-router-dom'
import * as GQL from '../../../../shared/src/graphql/schema'
import { isSettingsValid, SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { Settings } from '../../schema/settings.schema'
import { eventLogger } from '../../tracking/eventLogger'
import { FilterChip } from '../FilterChip'
import { submitSearch, toggleSearchFilter } from '../helpers'
import { PatternTypeProps } from '..'

interface Props extends SettingsCascadeProps, Pick<PatternTypeProps, 'patternType'> {
    history: H.History
    authenticatedUser: Pick<GQL.IUser, never> | null

    /**
     * The current query.
     */
    query: string
}

/**
 * A list of search scopes from the settings, shown as filter chips.
 */
export const SearchScopes: React.FunctionComponent<Props> = ({
    settingsCascade,
    query,
    authenticatedUser,
    history,
    patternType,
}) => {
    const scopes = (isSettingsValid<Settings>(settingsCascade) && settingsCascade.final['search.scopes']) || []

    const onSearchScopeClicked = useCallback(
        (value: string): void => {
            eventLogger.log('SearchScopeClicked', { search_filter: value })

            const newQuery = toggleSearchFilter(query, value)

            submitSearch(history, newQuery, 'filter', patternType, false)
        },
        [history, patternType, query]
    )

    return (
        <>
            {scopes
                .filter(scope => scope.value !== '') // clicking on empty scope would not trigger search
                .map((scope, i) => (
                    <FilterChip
                        query={query}
                        onFilterChosen={onSearchScopeClicked}
                        key={i}
                        value={scope.value}
                        name={scope.name}
                    />
                ))}
            {authenticatedUser && (
                <div className="search-scopes__edit">
                    <NavLink className="search-scopes__add-edit" to="/settings">
                        <small className="search-scopes__center">
                            {scopes.length === 0 ? (
                                <span className="search-scopes__add-scopes">Add search scopes for quick filtering</span>
                            ) : (
                                'Edit'
                            )}
                        </small>
                    </NavLink>
                </div>
            )}
        </>
    )
}
