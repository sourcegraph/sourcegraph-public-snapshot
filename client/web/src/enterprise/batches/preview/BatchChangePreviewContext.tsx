import React, { useState, useCallback } from 'react'

import { noop, uniqBy } from 'lodash'

import type {
    ChangesetSpecOperation,
    ChangesetSpecPublicationStateInput,
    ChangesetState,
} from '../../../graphql-operations'
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

type RecalculationRecord = [timestamp: number, status: 'pending' | 'complete']

export interface BatchChangePreviewContextState {
    // Filters are required to fetch all the changeset specs if all are selected
    // when publishing.
    readonly filters: BatchChangePreviewFilters
    setFilters: (filters: BatchChangePreviewFilters) => void
    // Be able to determine when the filter changes. This always having the context of know if all the visible
    // changesets have changed.
    readonly filtersChanged: boolean
    setFiltersChanged: (changed: boolean) => void
    // Maps any changesets to modified publish statuses set from the UI, to be included in
    // the mutation to apply the preview.
    readonly publicationStates: ChangesetSpecPublicationStateInput[]
    updatePublicationStates: (publicationStates: ChangesetSpecPublicationStateInput[]) => void
    // A list of tuples where each tuple represents a time preview publication states were
    // modified. The first element of the tuple is the timestamp (in number of ms) at
    // which the update occurred. We really only care about knowing the number of times we
    // have recalculated so far, but the timestamp gives us a stable key to use for
    // creating an array of React banner components from the updates. The second element
    // of the tuple is the status of requerying the `applyPreview` connection with the new
    // publication states. This allows us to hide the banner for an update until the
    // requery is complete.
    readonly recalculationUpdates: RecalculationRecord[]
    // Callback to mark all pending recalculation updates as complete, once our
    // `applyPreview` connection data is up-to-date with actions for the newest
    // publication states.
    resolveRecalculationUpdates: () => void
}

export const defaultState = (): BatchChangePreviewContextState => ({
    filters: defaultFilters(),
    setFilters: noop,
    filtersChanged: false,
    setFiltersChanged: noop,
    publicationStates: [],
    updatePublicationStates: noop,
    recalculationUpdates: [],
    resolveRecalculationUpdates: noop,
})

/**
 * A context tracking the filters for the Batch Changes preview page.
 *
 * @see BatchChangePreviewContextProvider
 */
export const BatchChangePreviewContext = React.createContext<BatchChangePreviewContextState>(defaultState())

export const BatchChangePreviewContextProvider: React.FunctionComponent<React.PropsWithChildren<{}>> = ({
    children,
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
    const [filtersChanged, setFiltersChanged] = useState(false)

    const [publicationStates, setPublicationStates] = useState<ChangesetSpecPublicationStateInput[]>([])

    // A list of tuples where each tuple represents a time preview publication states were
    // modified. The first element of the tuple is the timestamp (in number of ms) at
    // which the update occurred. We really only care about knowing the number of times we
    // have recalculated so far, but the timestamp gives us a stable key to use for
    // creating an array of React banner components from the updates. The second element
    // of the tuple is the status of requerying the `applyPreview` connection with the new
    // publication states. This allows us to hide the banner for an update until the
    // requery is complete.
    const [recalculationUpdates, setRecalculationUpdates] = useState<RecalculationRecord[]>([])
    const addRecalculationUpdate = useCallback((date: Date) => {
        setRecalculationUpdates(recalculationUpdates => [...recalculationUpdates, [date.getTime(), 'pending']])
    }, [])

    // Merge the new set of modified publication states with what's already been modified,
    // favoring the newest state set for a given changeset spec
    const updatePublicationStates = useCallback(
        (newPublicationStates: ChangesetSpecPublicationStateInput[]) => {
            // uniqBy removes duplicates by taking the first item it finds for a
            // `changesetSpec`, so we spread the updated publication states first so that
            // they get precedence
            setPublicationStates(uniqBy([...newPublicationStates, ...publicationStates], 'changesetSpec'))
            addRecalculationUpdate(new Date())
        },
        [publicationStates, addRecalculationUpdate]
    )

    // Callback to mark all pending recalculation updates as complete.
    const resolveRecalculationUpdates = useCallback(
        () =>
            setRecalculationUpdates(recalculationUpdates =>
                recalculationUpdates.map(([timestamp]) => [timestamp, 'complete'])
            ),
        []
    )

    return (
        <BatchChangePreviewContext.Provider
            value={{
                filters,
                setFilters,
                publicationStates,
                updatePublicationStates,
                recalculationUpdates,
                resolveRecalculationUpdates,
                filtersChanged,
                setFiltersChanged,
            }}
        >
            {children}
        </BatchChangePreviewContext.Provider>
    )
}
