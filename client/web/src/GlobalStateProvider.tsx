import { noop } from 'lodash';
import React, { Dispatch, SetStateAction, useCallback, useRef, useState as useReactState } from 'react';
import { createContext } from 'use-context-selector';

import { QueryState, SubmitSearchParameters, submitSearch as submitSearchInternal } from './search/helpers';

type QueryStateUpdate = QueryState | ((queryState: QueryState) => QueryState)
type SubmitQueryHandler =  (
    updateQuery: (currentQuery: string) => string,
    parameters: Omit<SubmitSearchParameters, 'query'>
) => void

export interface GlobalQueryState {
    queryState: QueryState
    setQueryState: (queryState: QueryStateUpdate) => void

    /**
     * submitSearch makes it possible to submit a new search query by updating
     * the current query via the callback.
     */
    submitSearch: (
        updateQuery: (currentQuery: string) => string,
        parameters: Omit<SubmitSearchParameters, 'query'>
    ) => void
}

/**
 * Main instance of global state context
 */
export const globalStateContext = createContext<GlobalQueryState>({
    queryState: { query: '' },
    setQueryState: noop,
    submitSearch: noop,
});

/**
 * Wrapper for native react useState hook, it provides additional API to be able
 * to get/pull state in memo function handlers without needs to put this state itself
 * in the memo dep list.
 *
 * Example:
 * ```
 *  const [state, setState, getState] = useState(initialState)
 *  // Note that we do not pass state itself but getState pull func that persist
 *  // the same between renders
 *  const handler = useCallback(() => doSomething(getState()), [getState])
 *
 * ```
 */
function useState<S>(initialState: S | (() => S)): [S, Dispatch<SetStateAction<S>>, () => S] {
    const [state, setState] = useReactState<S>(initialState)
    const stateReference = useRef<S>(state)
    stateReference.current = state

    const getState = useCallback(() => stateReference.current, [])

    return [state, setState, getState]
}

/**
 * Main Global state Provider (all shared state and logic)
 */
export const GlobalStateProvider: React.FunctionComponent = props => {
    const [queryState, setQueryState, getState] = useState<QueryState>({ query: '' })

    const submitSearch = useCallback<SubmitQueryHandler>((updateQuery, parameters) => {
        submitSearchInternal({
            ...parameters,
            query: updateQuery(getState().query)
        })
    }, [getState])

    return (
        <globalStateContext.Provider value={{ queryState, setQueryState, submitSearch }}>
            {props.children}
        </globalStateContext.Provider>
    )
}
