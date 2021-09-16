// NOTE (@fkling): The use of 'zustand' in this codebase should be considered as
// experimental until we had more time to evaluate this library. General
// application of this library is not recommended at this point.
// It is used here because it solves a very real performance issue
// (see https://github.com/sourcegraph/sourcegraph/issues/21200).
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
