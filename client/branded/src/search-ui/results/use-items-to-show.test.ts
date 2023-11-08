import { act, renderHook } from '@testing-library/react'
import { times } from 'lodash'
import { afterEach, describe, expect, it } from 'vitest'

import { INCREMENTAL_ITEMS_TO_SHOW, DEFAULT_INITIAL_ITEMS_TO_SHOW, useItemsToShow } from './use-items-to-show'

const RESULTS_NUMBER = 50

function renderUseItemsToShowHook(query = 'Hello there!') {
    return renderHook(() => useItemsToShow(query, RESULTS_NUMBER))
}

function scrollToViewMoreResults(scrollNumber: number, handleBottomHit: () => void) {
    // Do not await `act` call with sync logic. It's not a promise.

    act(() => {
        times(scrollNumber, handleBottomHit)
    })
}

describe('useItemsToShow', () => {
    afterEach(() => {
        sessionStorage.clear()
    })

    it('returns expected default values', () => {
        const { result } = renderUseItemsToShowHook()
        const { itemsToShow, handleBottomHit } = result.current

        expect(handleBottomHit).toEqual(expect.any(Function))
        expect(itemsToShow).toBe(DEFAULT_INITIAL_ITEMS_TO_SHOW)
    })

    it('increases `itemsToShow` value on `handleBottomHit()` call', () => {
        const { result: initialResult } = renderUseItemsToShowHook()
        const { handleBottomHit } = initialResult.current

        scrollToViewMoreResults(2, handleBottomHit)

        const { result } = renderUseItemsToShowHook()
        const { itemsToShow } = result.current

        expect(itemsToShow).toBe(DEFAULT_INITIAL_ITEMS_TO_SHOW + INCREMENTAL_ITEMS_TO_SHOW * 2)
    })

    it('preserves current `itemsToShow` value on a component unmount', () => {
        const { result: initialResult } = renderUseItemsToShowHook()
        const { handleBottomHit } = initialResult.current

        scrollToViewMoreResults(2, handleBottomHit)

        const { unmount } = renderUseItemsToShowHook()
        unmount()

        const { result } = renderUseItemsToShowHook()
        const { itemsToShow } = result.current

        expect(itemsToShow).toBe(DEFAULT_INITIAL_ITEMS_TO_SHOW + INCREMENTAL_ITEMS_TO_SHOW * 2)
    })

    it("doesn't go over the current results number", () => {
        const { result: initialResult } = renderUseItemsToShowHook()
        const { handleBottomHit } = initialResult.current

        scrollToViewMoreResults(20, handleBottomHit)

        const { result } = renderUseItemsToShowHook()
        const { itemsToShow } = result.current

        expect(itemsToShow).toBe(RESULTS_NUMBER)
    })

    it('resets `itemsToShow` if a new query is provided', () => {
        const { result: initialResult } = renderUseItemsToShowHook()
        const { handleBottomHit } = initialResult.current

        scrollToViewMoreResults(2, handleBottomHit)

        const { result } = renderUseItemsToShowHook('New query here!')
        const { itemsToShow } = result.current

        expect(itemsToShow).toBe(DEFAULT_INITIAL_ITEMS_TO_SHOW)
    })
})
