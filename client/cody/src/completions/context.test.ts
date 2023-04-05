import { bestJaccardMatch, jaccardScore, getWords } from './context'

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

describe('jaccardScore', () => {
    it('perfect overlap', () => {
        const targetWords = new Map<string, number>([
            ['foo', 1],
            ['bar', 1],
            ['baz', 1],
        ])
        const matchWords = new Map<string, number>([
            ['foo', 1],
            ['bar', 1],
            ['baz', 1],
        ])
        expect(jaccardScore(targetWords, matchWords)).toEqual(1)
    })
    it('no overlap', () => {
        const targetWords = new Map<string, number>([
            ['foo', 1],
            ['bar', 1],
            ['baz', 1],
        ])
        const matchWords = new Map<string, number>([
            ['qux', 1],
            ['quux', 1],
            ['quuz', 1],
        ])
        expect(jaccardScore(targetWords, matchWords)).toEqual(0)
    })
    it('50% overlap', () => {
        const targetWords = new Map<string, number>([
            ['bar', 1],
            ['baz', 1],
        ])
        const matchWords = new Map<string, number>([
            ['foo', 1],
            ['bar', 1],
            ['baz', 1],
            ['qux', 1],
        ])
        expect(jaccardScore(targetWords, matchWords)).toEqual(0.5)
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
