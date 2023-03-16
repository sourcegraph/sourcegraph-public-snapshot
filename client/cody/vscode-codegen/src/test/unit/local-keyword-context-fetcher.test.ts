import assert from 'assert'

import { getTermScore } from '../../keyword-context'

describe('getTermScore', () => {
    it('generate term score based on letter cases', () => {
        const testCases = [
            // Test with a lowercase term
            { input: 'sourcegraph', expectedOutput: 'sourcegraph'.length },
            // Test with an uppercase term
            { input: 'SOURCEGRAPH', expectedOutput: 'SOURCEGRAPH'.length },
            // Test with a mixed case term
            { input: 'useSourcegraph', expectedOutput: 'useSourcegraph'.length * 10 },
            // Test with a term that contains no letters
            { input: '123', expectedOutput: '123'.length },
            // Test with an empty term
            { input: '', expectedOutput: ''.length },
        ]

        testCases.forEach(({ input, expectedOutput }) => {
            it(`generates term score based on letter cases for input "${input}"`, () => {
                assert.deepStrictEqual(getTermScore(input), expectedOutput)
            })
        })
    })
})
