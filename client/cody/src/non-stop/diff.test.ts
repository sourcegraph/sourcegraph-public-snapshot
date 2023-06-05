import { computeDiff, dumpUse, longestCommonSubsequence } from './diff'

// Note, computeDiff does not treat its arguments symmetrically
// because we present *Cody's* edits as the "foreign" ones an the
// human edits as benign. The arguments are:
// computeDiff(original text, cody text, human text, ...)
describe('computeDiff', () => {
    it('should merge trivial Cody insertions', () => {
        const diff = computeDiff('hello, world!', 'hello, worldly humans of earth!', 'hello, world!', {
            line: 42,
            character: 7,
        })
        expect(diff.clean).toBe(true)
        expect(diff.conflicts).toStrictEqual([])
        expect(diff.edits).toStrictEqual([
            {
                kind: 'insert',
                text: 'ly humans of earth',
                range: {
                    start: { line: 42, character: 7 + 'hello, world'.length },
                    end: { line: 42, character: 7 + 'hello, world'.length },
                },
            },
        ])
    })
    it('should merge non-overlapping human and Cody insertions', () => {
        const diff = computeDiff('hello, world!', 'hello, worldly humans of earth!', 'hello, salutations world!', {
            line: 42,
            character: 7,
        })
        expect(diff.clean).toBe(true)
        expect(diff.conflicts).toStrictEqual([])
        expect(diff.edits).toStrictEqual([
            {
                kind: 'insert',
                text: 'ly humans of earth',
                range: {
                    start: { line: 42, character: 7 + 'hello, salutations world'.length },
                    end: { line: 42, character: 7 + 'hello, salutations world'.length },
                },
            },
        ])
    })
    it('should merge overlapping, identical human and Cody edits', () => {
        const diff = computeDiff('hello, world!', 'hello, puny earthlings', 'hello, puny earthlings', {
            line: 42,
            character: 7,
        })
        expect(diff.clean).toBe(true)
        expect(diff.conflicts).toStrictEqual([])
        expect(diff.edits).toStrictEqual([])
    })
    it('should report conflicts', () => {
        const diff = computeDiff('hello, world!', 'hello, WORLD!', 'hello, earth!', { line: 42, character: 7 })
        expect(diff.clean).toBe(false)
    })
})

describe('longestCommonSubsequence', () => {
    it('identical strings should use themselves', () => {
        const palindrome = 'amanaplanacanalpanama'
        const lcs = longestCommonSubsequence(palindrome, palindrome)
        // Because the strings are identical, this should be a diagonal
        // matrix indicating every character is used.
        dumpUse(lcs, palindrome, palindrome)
        for (let v = 0; v < palindrome.length; v++) {
            for (let u = 0; u < palindrome.length; u++) {
                expect(lcs[(v + 1) * (palindrome.length + 1) + (u + 1)]).toBe(u === v ? 1 : 0)
            }
        }
    })
    it('prefixes should use the prefix', () => {
        const prefix = 'hello, '
        const a = prefix + 'world!'
        const b = prefix + 'peeps...'
        const lcs = longestCommonSubsequence(a, b)
        for (let v = 0; v < b.length; v++) {
            for (let u = 0; u < a.length; u++) {
                const entry = lcs[(v + 1) * (a.length + 1) + (u + 1)]
                if (u === v && u < prefix.length) {
                    // Because the prefix is identical, the prefix
                    // should be used in its entirety.
                    expect(entry).toBe(1)
                } else {
                    // There is nothing else in common.
                    expect(entry).toBe(0)
                }
            }
        }
    })
    it('subsequences should be used', () => {
        const a = '.a...bc....d...e...'
        const b = '~~~ab~cde~~~'
        const lcs = longestCommonSubsequence(a, b)
        dumpUse(lcs, a, b)
        for (let v = 0; v < b.length; v++) {
            for (let u = 0; u < a.length; u++) {
                const entry = lcs[(v + 1) * (a.length + 1) + (u + 1)]
                // Note, this condition is not true *in general*, but
                // because of the way these inputs are constructed:
                // a and b have a common subsequence, abcde, and
                // nothing else in common
                expect(entry).toBe(a[u] === b[v] ? 1 : 0)
            }
        }
    })
})
