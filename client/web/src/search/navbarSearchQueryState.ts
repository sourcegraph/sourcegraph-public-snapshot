import create from 'zustand'

import { QueryState } from './helpers'

interface NavbarQueryState {
    queryState: QueryState
    setQueryState: (queryState: QueryState) => void
}
export const useNavbarQueryState = create<NavbarQueryState>(set => ({
    queryState: { query: '' },
    setQueryState: queryState => set({ queryState }),
}))
