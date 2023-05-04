import { sliceUntilFirstNLinesOfSuffixMatch } from './provider'

describe('sliceUntilFirstNLinesOfSuffixMatch', () => {
    it('returns the right text', () => {
        const suggestion = 'foo\nbar\nbaz\nline-1\nline-2\noh no\nline-1\nline-2\nline-3'
        const suffix = 'line-1\nline-2\nline-3\nline-4\nline-5'

        expect(sliceUntilFirstNLinesOfSuffixMatch(suggestion, suffix, 3)).toMatchInlineSnapshot(`
            "foo
            bar
            baz
            line-1
            line-2
            oh no"
        `)
    })

    it('works with the example suggested by Cody', () => {
        const suggestion = 'foo\nbar\nbaz\nqux\nquux'
        const suffix = 'baz\nqux\nquux'

        expect(sliceUntilFirstNLinesOfSuffixMatch(suggestion, suffix, 3)).toMatchInlineSnapshot(`
            "foo
            bar"
        `)
    })
})
