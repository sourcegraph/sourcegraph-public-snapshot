import { useCallback, useEffect, useState } from 'react'

import { useLocation, useNavigate } from 'react-router-dom'

import type { LegacyBatchChangesFilter } from '@sourcegraph/shared/src/settings/temporary/TemporarySettings'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'

import { BatchChangeState } from '../../../graphql-operations'

const STATUS_OPTIONS = [BatchChangeState.OPEN, BatchChangeState.DRAFT, BatchChangeState.CLOSED]

// Drafts are a new feature of serverside execution that for now should not be shown if
// execution is not enabled.
const STATUS_OPTIONS_NO_DRAFTS: BatchChangeState[] = [BatchChangeState.OPEN, BatchChangeState.CLOSED]

const fromLegacyFilters = (legacyFilters: LegacyBatchChangesFilter[]): BatchChangeState[] =>
    legacyFilters.map(legacyFilter => legacyFilter.value)

const toLegacyFilters = (filters: BatchChangeState[]): LegacyBatchChangesFilter[] =>
    filters.map(filter => ({ label: filter.toString(), value: filter }))

interface UseBatchChangeListFiltersProps {
    isExecutionEnabled: boolean
}

interface UseBatchChangeListFiltersResult {
    /** State representing the different filters selected in the `MultiSelect` UI. */
    selectedFilters: BatchChangeState[]

    /** Method to set the filters selected in the `MultiSelect`. */
    setSelectedFilters: (filters: BatchChangeState[]) => void

    /**
     * List of available batch changes filters, it may be different based on
     * {@link isExecutionEnabled} prop
     */
    availableFilters: BatchChangeState[]
}

/**
 * Custom hook for managing and persisting filter options selected from
 * the MultiCombobox UI to filter a list of batch changes.
 */
export const useBatchChangeListFilters = (props: UseBatchChangeListFiltersProps): UseBatchChangeListFiltersResult => {
    const { isExecutionEnabled } = props

    const navigate = useNavigate()
    const location = useLocation()

    const availableFilters = isExecutionEnabled ? STATUS_OPTIONS : STATUS_OPTIONS_NO_DRAFTS

    // NOTE: Fetching this setting is an async operation, so we can't use it as the
    // initial value for `useState`. Instead, we will set the value of the filter state in
    // a `useEffect` hook once we've loaded it.
    const [defaultFilters, setDefaultFilters] = useTemporarySetting('batches.defaultListFilters', [])

    const [hasModifiedFilters, setHasModifiedFilters] = useState(false)
    const [selectedFilters, setSelectedFiltersRaw] = useState<BatchChangeState[]>(() => {
        const searchParameters = new URLSearchParams(location.search).get('states')

        if (searchParameters) {
            const loweredCaseFilters = new Set(availableFilters.map(filter => filter.toLowerCase()))
            const urlFilters = searchParameters.split(',').map(option => option.toLowerCase())

            return urlFilters
                .filter(urlFilter => loweredCaseFilters.has(urlFilter))
                .map(urlFilters => urlFilters.toUpperCase() as BatchChangeState)
        }

        return []
    })

    const setSelectedFilters = useCallback(
        (filters: BatchChangeState[]) => {
            const searchParameters = new URLSearchParams(location.search)

            if (filters.length > 0) {
                searchParameters.set(
                    'states',
                    filters
                        .map(filter => filter.toLowerCase())
                        .join(',')
                        .toLowerCase()
                )
            } else {
                searchParameters.delete('states')
            }

            if (location.search !== searchParameters.toString()) {
                navigate({ search: searchParameters.toString() }, { replace: true })
            }

            setHasModifiedFilters(true)
            setSelectedFiltersRaw(filters)
            setDefaultFilters(toLegacyFilters(filters))
        },
        [setDefaultFilters, navigate, location.search]
    )

    // Once we've loaded the default filters from temporary settings, we will set them in
    // state, but if the user has already modified the filters before then, or we read a
    // different set of filters from the URL search params, those will take precedence.
    useEffect(() => {
        const searchParameters = new URLSearchParams(location.search).get('states')

        if (defaultFilters && !hasModifiedFilters && !searchParameters) {
            setSelectedFiltersRaw(fromLegacyFilters(defaultFilters))
        }
    }, [defaultFilters, hasModifiedFilters, location.search])

    return {
        selectedFilters,
        setSelectedFilters,
        availableFilters,
    }
}
