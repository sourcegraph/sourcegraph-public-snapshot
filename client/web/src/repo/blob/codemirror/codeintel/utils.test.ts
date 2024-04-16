import { describe, expect, it } from 'vitest'

import { syncValues } from './utils'

describe('blob/codemirror/codeintel/utils', () => {
    describe('syncValues', () => {
        const create = (n: number) => n * 2
        const update = (n: number) => n + 1

        it('returns the current list of values if the inputs did not change', () => {
            const values = [2, 4, 6]
            const previousInput = [1, 2, 3]
            const currentInput = previousInput
            expect(syncValues({ values, previousInput, currentInput, create, update: n => n })).toBe(values)
        })

        it('calls create for new input values and update for existing input values', () => {
            const values = [2, 6]
            const previousInput = [1, 3]
            const currentInput = [1, 2, 3]
            expect(syncValues({ values, previousInput, currentInput, create, update })).toEqual([3, 4, 7])
        })

        it('removes values not present in the current iput', () => {
            const values = [2, 8]
            const previousInput = [1, 4]
            const currentInput = [1]
            expect(syncValues({ values, previousInput, currentInput, create, update })).toEqual([3])
        })
    })
})
