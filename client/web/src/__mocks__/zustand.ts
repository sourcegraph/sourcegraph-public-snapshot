// Taken from https://github.com/pmndrs/zustand/wiki/Testing

import actualCreate, { StateCreator } from 'zustand'
import { act } from 'react-dom/test-utils'

// a variable to hold reset functions for all stores declared in the app
const storeResetFns: Set<() => void> = new Set()

// when creating a store, we get its initial state, create a reset function and add it in the set
const create = <T extends object>(createState: StateCreator<T>) => {
    const store = actualCreate(createState)
    const initialState = store.getState()
    storeResetFns.add(() => store.setState(initialState, true))
    return store
}

// Reset all stores after each test run
afterEach(() => {
    act(() => storeResetFns.forEach(resetFn => resetFn()))
})

export default create
