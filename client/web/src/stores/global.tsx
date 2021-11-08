import React from 'react'
import create from 'zustand'
import createContext from 'zustand/context'

import { NavbarQueryState, createNavbarQueryStateStore } from './navbarSearchQueryState'

interface GlobalStore extends NavbarQueryState {}

const { Provider, useStore } = createContext<GlobalStore>()
export { useStore as useGlobalStore }

export const GlobalStoreProvider: React.FunctionComponent<{
    initialStore?: Partial<GlobalStore>
    children: React.ReactChild | React.ReactChildren
}> = ({ initialStore, children }) => (
    <Provider
        createStore={() =>
            create((set, get) => ({
                ...createNavbarQueryStateStore(set, get),
                ...initialStore,
            }))
        }
    >
        {children}
    </Provider>
)
