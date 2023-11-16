import { describe, expect, it } from 'vitest'

import { getListWindow } from './get-list-window'

describe('getListWindow', () => {
    it('should return unmodified list in case if list has less element than size', () => {
        expect(getListWindow([1, 2, 3, 4, 5], 2, 10)).toStrictEqual({
            window: [1, 2, 3, 4, 5],
            leftRemaining: 0,
            rightRemaining: 0,
        })
    })

    it('should return correct window if list has enough items from both sides around pivot index', () => {
        expect(getListWindow([1, 2, 3, 4, 5], 2, 3)).toStrictEqual({
            window: [2, 3, 4],
            leftRemaining: 1,
            rightRemaining: 1,
        })
    })

    it('should return correct window if list has enough items only at right side around pivot index', () => {
        expect(getListWindow([1, 2, 3, 4, 5], 1, 4)).toStrictEqual({
            window: [1, 2, 3, 4],
            leftRemaining: 0,
            rightRemaining: 1,
        })
    })

    it('should return correct window if list has enough items only at left side around pivot index', () => {
        expect(getListWindow([1, 2, 3, 4, 5], 3, 4)).toStrictEqual({
            window: [2, 3, 4, 5],
            leftRemaining: 1,
            rightRemaining: 0,
        })
    })
})
