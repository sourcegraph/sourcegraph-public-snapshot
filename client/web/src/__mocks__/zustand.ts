// Taken from https://github.com/pmndrs/zustand/wiki/Testing

import { afterAll, afterEach } from '@jest/globals'
import type { Act } from '@testing-library/react-hooks'
import { act } from 'react-dom/test-utils'
import actualCreate, { type StateCreator, type UseStore } from 'zustand'

// This allows test suites to specify which 'act' funtion to use. This is
// necessary if test suites use a different renderer.
let actToUse: Act = act

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
    actToUse(() => {
        for (const resetFunc of storeResetFns) {
            resetFunc()
        }
    })
})

afterAll(() => {
    actToUse = act
})

/**
 * Use this function to overwrite the 'act' function that should be used be
 * this mock to reset stores. Setting this is necessary if you use a renderer
 * different from 'react-dom/test-utils'.
 */
export function setAct(newAct: Act): void {
    actToUse = newAct
}

// eslint-disable-next-line import/no-default-export
export default create
