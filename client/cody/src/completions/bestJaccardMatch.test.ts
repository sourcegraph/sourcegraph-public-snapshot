import { bestJaccardMatch, getWords } from './bestJaccardMatch'

const targetSnippet = `
import { bestJaccardMatch, getWords } from './context'

describe('getWords', () => {
    it('works with regular text', () => {
        expect(getWords('foo bar baz')).toEqual(
            new Map<string, number>([
                ['foo', 1],
                ['bar', 1],
                ['baz', 1],
            ])
        )
        expect(getWords('running rocks slipped over')).toEqual(
            new Map<string, number>([
                ['run', 1],
                ['rock', 1],
                ['slip', 1],
            ])
        )
    })
})
`

const matchSnippet = `
describe('bestJaccardMatch', () => {
    it('should return the best match', () => {
        const matchText = [
            'foo',
            'bar',
            'baz',
            'qux',
            'quux',
        ].join('\n')
    })
})
`

describe('getWords', () => {
    it('works with regular text', () => {
        expect(getWords('foo bar baz')).toEqual(
            new Map<string, number>([
                ['foo', 1],
                ['bar', 1],
                ['baz', 1],
            ])
        )
        expect(getWords('running rocks slipped over')).toEqual(
            new Map<string, number>([
                ['run', 1],
                ['rock', 1],
                ['slip', 1],
            ])
        )
    })

    it('works with code snippets', () => {
        expect(getWords(targetSnippet)).toEqual(
            new Map<string, number>([
                ['import', 1],
                ['bestjaccardmatch', 1],
                ['getword', 4],
                ['context', 1],
                ['describ', 1],
                ['work', 1],
                ['regular', 1],
                ['text', 1],
                ['expect', 2],
                ['foo', 2],
                ['bar', 2],
                ['baz', 2],
                ['toequal', 2],
                ['new', 2],
                ['map', 2],
                ['string', 2],
                ['number', 2],
                ['1', 6],
                ['run', 2],
                ['rock', 2],
                ['slip', 2],
            ])
        )
    })
})

describe('bestJaccardMatch', () => {
    it('should return the best match', () => {
        const matchText = [
            'foo',
            'bar',
            'baz',
            'qux',
            'quux',
            'quuz',
            'corge',
            'grault',
            'garply',
            'waldo',
            'fred',
            'plugh',
            'xyzzy',
            'thud',
        ].join('\n')
        expect(bestJaccardMatch('foo\nbar\nbaz', matchText, 3)).toEqual({
            score: 1,
            content: 'foo\nbar\nbaz',
        })
        expect(bestJaccardMatch('bar\nquux', matchText, 4)).toEqual({
            score: 0.5,
            content: 'bar\nbaz\nqux\nquux',
        })
        expect(
            bestJaccardMatch(
                ['grault', 'notexist', 'garply', 'notexist', 'waldo', 'notexist', 'notexist'].join('\n'),
                matchText,
                6
            )
        ).toEqual({
            score: 0.3,
            content: ['quux', 'quuz', 'corge', 'grault', 'garply', 'waldo'].join('\n'),
        })
    })

    it('works with code snippets', () => {
        expect(bestJaccardMatch(targetSnippet, matchSnippet, 5)).toMatchInlineSnapshot(`
            Object {
              "content": "describe('bestJaccardMatch', () => {
                it('should return the best match', () => {
                    const matchText = [
                        'foo',
                        'bar',",
              "score": 0.08695652173913043,
            }
        `)
    })
})
