import React, { useState } from 'react'

import { ChangesetSpecOperation, ChangesetState } from '../../../graphql-operations'
import { isValidChangesetSpecOperation, isValidChangesetState } from '../utils'

export interface BatchChangePreviewFilters {
    search: string | null
    currentState: ChangesetState | null
    action: ChangesetSpecOperation | null
}

const defaultFilters = (): BatchChangePreviewFilters => ({
    search: null,
    currentState: null,
    action: null,
})

export interface BatchChangePreviewPagination {
    first: number | null
    after: string | null
}

const defaultPagination = (): BatchChangePreviewPagination => ({
    first: null,
    after: null,
})

export interface BatchChangePreviewContextState {
    // Filters are required to fetch all the changeset specs if all are selected
    // when publishing.
    filters: BatchChangePreviewFilters
    setFilters: (filters: BatchChangePreviewFilters) => void

    // Pagination is required to fetch all the changeset specs if all are
    // selected when publishing.
    pagination: BatchChangePreviewPagination
    setPagination: (pagination: BatchChangePreviewPagination) => void

    // We need to track if there are more pages than are currently visible: if
    // so, then the UI to select all changesets must be distinct from selecting
    // visible changesets.
    hasMorePages: boolean
    setHasMorePages: (hasMorePages: boolean) => void

    // It's also helpful to know how many total changesets there are.
    totalCount: number
    setTotalCount: (totalCount: number) => void
}

export const defaultState = (): BatchChangePreviewContextState => ({
    filters: defaultFilters(),
    setFilters: () => {},
    pagination: defaultPagination(),
    setPagination: () => {},
    hasMorePages: false,
    setHasMorePages: () => {},
    totalCount: 0,
    setTotalCount: () => {},
})

/**
 * A context tracking the filters, pagination, and other miscellany for an
 * instance of the Batch Changes preview page.
 *
 * Generally, this context will own the filter and connection state, but NOT the
 * pagination state, as that has to be owned by FilteredConnection. Use
 * BatchChangePreviewContextProvider to instantiate a context that owns the
 * filter state.
 *
 * @see BatchChangePreviewContextProvider
 */
export const BatchChangePreviewContext = React.createContext<BatchChangePreviewContextState>(defaultState())

export const BatchChangePreviewContextProvider: React.FunctionComponent<{ initialHasMorePages?: boolean }> = ({
    children,
    initialHasMorePages,
}) => {
    const urlParameters = new URLSearchParams(location.search)

    const [filters, setFilters] = useState<BatchChangePreviewFilters>(() => {
        const action = urlParameters.get('action')
        const currentState = urlParameters.get('current_state')
        const search = urlParameters.get('search')

        return {
            action: action && isValidChangesetSpecOperation(action) ? action : null,
            currentState: currentState && isValidChangesetState(currentState) ? currentState : null,
            search: search ?? null,
        }
    })

    const [pagination, setPagination] = useState(defaultPagination())
    const [hasMorePages, setHasMorePages] = useState(!!initialHasMorePages)
    const [totalCount, setTotalCount] = useState(0)

    return (
        <BatchChangePreviewContext.Provider
            value={{
                filters,
                setFilters,
                pagination,
                setPagination,
                hasMorePages,
                setHasMorePages,
                totalCount,
                setTotalCount,
            }}
        >
            {children}
        </BatchChangePreviewContext.Provider>
    )
}
