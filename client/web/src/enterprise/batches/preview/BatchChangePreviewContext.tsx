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
    filters: BatchChangePreviewFilters
    setFilters: (filters: BatchChangePreviewFilters) => void

    pagination: BatchChangePreviewPagination
    setPagination: (pagination: BatchChangePreviewPagination) => void
}

const defaultState = (): BatchChangePreviewContextState => ({
    filters: defaultFilters(),
    setFilters: () => {},
    pagination: defaultPagination(),
    setPagination: () => {},
})

/**
 * A context tracking the filters and pagination for an instance of the Batch
 * Changes preview page.
 *
 * Generally, this context will own the filter state, but NOT the pagination
 * state, as that has to be owned by FilteredConnection. Use
 * BatchChangePreviewContextProvider to instantiate a context that owns the
 * filter state.
 *
 * @see BatchChangePreviewContextProvider
 */
export const BatchChangePreviewContext = React.createContext<BatchChangePreviewContextState>(defaultState())

export const BatchChangePreviewContextProvider: React.FunctionComponent<{}> = ({ children }) => {
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

    const [pagination, setPagination] = useState<BatchChangePreviewPagination>(defaultPagination())

    return (
        <BatchChangePreviewContext.Provider
            value={{
                filters,
                setFilters,
                pagination,
                setPagination,
            }}
        >
            {children}
        </BatchChangePreviewContext.Provider>
    )
}
