import { useCallback, useEffect, useMemo, useState } from 'react'

import { lowerCase } from 'lodash'
import { useHistory } from 'react-router'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { MultiSelectState } from '@sourcegraph/wildcard'

import { BatchChangeState } from '../../../graphql-operations'

import { STATUS_OPTIONS } from './BatchChangeListFilters'

const statesToFilters = (states: string[]): MultiSelectState<BatchChangeState> =>
    STATUS_OPTIONS.filter(option => states.map(lowerCase).includes(option.value.toLowerCase()))

const filtersToStates = (filters: MultiSelectState<BatchChangeState>): BatchChangeState[] =>
    filters.map(filter => filter.value)

interface UseBatchChangeListFiltersResult {
    /** State representing the different filters selected in the `MultiSelect` UI. */
    selectedFilters: MultiSelectState<BatchChangeState>
    /** Method to set the filters selected in the `MultiSelect`. */
    setSelectedFilters: (filters: MultiSelectState<BatchChangeState>) => void
    /**
     * Array of raw `BatchChangeState`s corresponding to `selectedFilters`, i.e. for
     * passing in GraphQL connection query parameters.
     */
    selectedStates: BatchChangeState[]
}

/**
 * Custom hook for managing, persisting, and transforming the state options selected from
 * the `MultiSelect` UI to filter a list of batch changes.
 */
export const useBatchChangeListFilters = (): UseBatchChangeListFiltersResult => {
    const history = useHistory()

    // NOTE: Fetching this setting is an async operation, so we can't use it as the
    // initial value for `useState`. Instead, we will set the value of the filter state in
    // a `useEffect` hook once we've loaded it.
    const [defaultFilters, setDefaultFilters] = useTemporarySetting('batches.defaultListFilters', [])

    const [selectedFilters, setSelectedFiltersRaw] = useState<MultiSelectState<BatchChangeState>>(() => {
        const searchParameters = new URLSearchParams(history.location.search).get('states')
        if (searchParameters) {
            return statesToFilters(searchParameters.split(','))
        }
        return []
    })

    const [hasModifiedFilters, setHasModifiedFilters] = useState(false)

    const setSelectedFilters = useCallback(
        (filters: MultiSelectState<BatchChangeState>) => {
            setHasModifiedFilters(true)
            setSelectedFiltersRaw(filters)
            setDefaultFilters(filters)

            const searchParameters = new URLSearchParams(history.location.search)
            if (filters.length > 0) {
                searchParameters.set('states', filtersToStates(filters).join(',').toLowerCase())
            } else {
                searchParameters.delete('states')
            }

            if (history.location.search !== searchParameters.toString()) {
                history.replace({ ...history.location, search: searchParameters.toString() })
            }
        },
        [setDefaultFilters, history]
    )

    // Once we've loaded the default filters from temporary settings, we will set them in
    // state, but if the user has already modified the filters before then, or we read a
    // different set of filters from the URL search params, those will take precedence.
    useEffect(() => {
        const searchParameters = new URLSearchParams(history.location.search).get('states')

        if (defaultFilters && !hasModifiedFilters && !searchParameters) {
            setSelectedFiltersRaw(defaultFilters)
        }
    }, [defaultFilters, hasModifiedFilters, history.location.search])

    const selectedStates = useMemo<BatchChangeState[]>(() => filtersToStates(selectedFilters), [selectedFilters])

    return {
        selectedFilters,
        setSelectedFilters,
        selectedStates,
    }
}
