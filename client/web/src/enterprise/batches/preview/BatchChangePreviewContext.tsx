import { noop } from 'lodash'
import React, { useState } from 'react'

import { ChangesetSpecOperation, ChangesetSpecPublicationStateInput, ChangesetState } from '../../../graphql-operations'
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

export interface BatchChangePreviewContextState {
    // Filters are required to fetch all the changeset specs if all are selected
    // when publishing.
    readonly filters: BatchChangePreviewFilters
    setFilters: (filters: BatchChangePreviewFilters) => void
    // Maps any changesets to modified publish statuses set from the UI, to be included in
    // the mutation to apply the preview.
    readonly publicationStates: ChangesetSpecPublicationStateInput[]
    setPublicationStates: (publicationStates: ChangesetSpecPublicationStateInput[]) => void
}

export const defaultState = (): BatchChangePreviewContextState => ({
    filters: defaultFilters(),
    setFilters: noop,
    publicationStates: [],
    setPublicationStates: noop,
})

/**
 * A context tracking the filters for the Batch Changes preview page.
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

    const [publicationStates, setPublicationStates] = useState<ChangesetSpecPublicationStateInput[]>([])

    return (
        <BatchChangePreviewContext.Provider
            value={{
                filters,
                setFilters,
                publicationStates,
                setPublicationStates,
            }}
        >
            {children}
        </BatchChangePreviewContext.Provider>
    )
}
