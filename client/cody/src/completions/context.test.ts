import { bestJaccardMatch, getWords } from './context'

describe('getWords', () => {
    it('', () => {
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
            text: 'foo\nbar\nbaz',
        })
        expect(bestJaccardMatch('bar\nquux', matchText, 4)).toEqual({
            score: 0.5,
            text: 'bar\nbaz\nqux\nquux',
        })
        expect(
            bestJaccardMatch(
                ['grault', 'notexist', 'garply', 'notexist', 'waldo', 'notexist', 'notexist'].join('\n'),
                matchText,
                6
            )
        ).toEqual({
            score: 0.3,
            text: ['quux', 'quuz', 'corge', 'grault', 'garply', 'waldo'].join('\n'),
        })
    })
})
