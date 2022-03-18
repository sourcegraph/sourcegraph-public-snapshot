import { useCallback, useEffect, useMemo, useState } from 'react'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { MultiSelectState } from '@sourcegraph/wildcard'

import { BatchChangeState } from '../../../graphql-operations'

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
    // NOTE: Fetching this setting is an async operation, so we can't use it as the
    // initial value for `useState`. Instead, we will set the value of the filter state in
    // a `useEffect` hook once we've loaded it, but only if the user has not changed the
    // filter state in the meantime.
    const [defaultFilters, setDefaultFilters] = useTemporarySetting('batches.defaultListFilters', [])

    const [selectedFilters, setSelectedFiltersRaw] = useState<MultiSelectState<BatchChangeState>>([])
    const [hasModifiedFilters, setHasModifiedFilters] = useState(false)

    const setSelectedFilters = useCallback(
        (filters: MultiSelectState<BatchChangeState>) => {
            setHasModifiedFilters(true)
            setSelectedFiltersRaw(filters)
            setDefaultFilters(filters)
        },
        [setDefaultFilters]
    )

    useEffect(() => {
        if (defaultFilters && !hasModifiedFilters) {
            setSelectedFiltersRaw(defaultFilters)
        }
    }, [defaultFilters, hasModifiedFilters])

    const selectedStates = useMemo<BatchChangeState[]>(() => selectedFilters.map(filter => filter.value), [
        selectedFilters,
    ])

    return {
        selectedFilters,
        setSelectedFilters,
        selectedStates,
    }
}
