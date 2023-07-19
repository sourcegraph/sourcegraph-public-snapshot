import { get, writable } from 'svelte/store'
import { describe, it, expect } from 'vitest'

import { createForwardStore } from './stores'

describe('createForwardStore', () => {
    it('syncs the initial value of the passed store', () => {
        const store = createForwardStore(writable(1))
        expect(get(store)).toBe(1)
    })

    it('updates when the passed store updates', () => {
        const origin = writable(1)
        const store = createForwardStore(origin)

        origin.set(2)

        expect(get(store)).toBe(2)
    })

    it('syncs with a new store', () => {
        const origin = writable(1)
        const store = createForwardStore(origin)
        store.updateStore(writable(2))

        // Update the original store to verify that the forward store doesn't
        // subscribe to it anymore
        origin.set(3)

        expect(get(store)).toBe(2)
    })

    it('updates the original store', () => {
        const origin = writable(1)
        const store = createForwardStore(origin)

        store.set(2)

        expect(get(origin)).toBe(2)
    })
})
