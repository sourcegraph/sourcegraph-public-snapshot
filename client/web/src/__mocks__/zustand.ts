// Taken from https://github.com/pmndrs/zustand/wiki/Testing

import { act } from 'react-dom/test-utils'
import actualCreate, { StateCreator, UseStore } from 'zustand'

// a variable to hold reset functions for all stores declared in the app
const storeResetFns: Set<() => void> = new Set()

// when creating a store, we get its initial state, create a reset function and add it in the set
const create = <T extends object>(createState: StateCreator<T>): UseStore<T> => {
    const store = actualCreate(createState)
    const initialState = store.getState()
    storeResetFns.add(() => store.setState(initialState, true))
    return store
}

// Reset all stores after each test run
afterEach(() => {
    act(() => {
        for (const resetFn of storeResetFns) {
            resetFn()
        }
    })
})

// eslint-disable-next-line import/no-default-export
export default create
