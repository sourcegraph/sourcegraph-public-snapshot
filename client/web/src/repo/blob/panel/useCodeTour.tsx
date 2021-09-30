import React, { useEffect, useMemo, useReducer } from 'react'
import { useHistory } from 'react-router'

import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'

// import { useTemporarySetting } from '../../../settings/temporary/useTemporarySetting'

import { Action, initialState, reducer, State, Routes } from './reducer'

export interface CodeTourContext {
    state: State
    dispatch: React.Dispatch<Action>
}

export const CodeTourContext = React.createContext<CodeTourContext | null>(null)
CodeTourContext.displayName = 'CodeTourContext'

export const useCodeTourContext = (): CodeTourContext => {
    const context = React.useContext(CodeTourContext)
    if (!context) {
        throw new Error('You are trying to use this component outside the CodeTourContext provider')
    }
    return context
}

export const CodeTourProvider: React.FunctionComponent = ({ children }) => {
    const history = useHistory()
    const pathname = `${history.location.pathname}${history.location.search}`

    const routes: Routes = {
        [pathname]: {
            index: pathname,
            description: '',
            lineNumber: history.location.search.slice(1),
        },
    }

    const [codeTour] = useLocalStorage('codetour.routeslist', { current: pathname, routes })
    const [state, dispatch] = useReducer(reducer, initialState(codeTour || { current: pathname, routes }))
    const contextValue = useMemo(() => ({ state, dispatch }), [state, dispatch])

    return <CodeTourContext.Provider value={contextValue}>{children}</CodeTourContext.Provider>
}

interface UseCodeTour {
    routes: Routes
    routesList: { index: string; description: string; lineNumber: string }[]
    currentRoute: string
    setRoute: () => void
    removeRoute: (index: string) => void
    updateDescription: (index: string, description: string) => void
}

export const useCodeTour = (): UseCodeTour => {
    const history = useHistory()
    const context = useCodeTourContext()
    const { state, dispatch } = context
    const [, setCodeTour] = useLocalStorage('codetour.routeslist', state)

    const pathname = history.location.pathname
    const search = history.location.search

    const getters = {
        routes: state.routes,
        routesList: Object.keys(state.routes).map(key => state.routes[key]),
        currentRoute: state.current,
    }

    useEffect(() => {
        setCodeTour(state)
    }, [setCodeTour, state])

    const lineNumber = history.location.search.slice(1)

    useMemo(() => {
        dispatch({
            type: 'SET_CURRENT_ROUTE',
            payload: {
                index: `${pathname}${search}`,
            },
        })
    }, [dispatch, pathname, search])

    const setters = useMemo(
        () => ({
            setRoute: (): void =>
                dispatch({
                    type: 'ADD_ROUTE',
                    payload: {
                        index: `${pathname}${search}`,
                        description: 'test',
                        lineNumber,
                    },
                }),
            removeRoute: (index: string): void => {
                const routes = { ...state.routes }
                delete routes[index]
                return dispatch({
                    type: 'REMOVE_ROUTE',
                    payload: {
                        routes,
                    },
                })
            },
            updateDescription: (index: string, description: string): void =>
                dispatch({
                    type: 'UPDATE_DESCRIPTION',
                    payload: {
                        index,
                        description,
                    },
                }),
        }),
        [dispatch, lineNumber, pathname, search, state.routes]
    )

    return { ...setters, ...getters }
}
