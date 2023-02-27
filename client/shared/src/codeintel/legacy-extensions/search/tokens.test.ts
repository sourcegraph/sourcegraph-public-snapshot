import * as assert from 'assert'

import { cStyleBlockComment, slashPattern } from '../language-specs/comments'

import { findSearchToken } from './tokens'

describe('findSearchToken', () => {
    it('custom identCharPattern', () => {
        assert.deepStrictEqual(
            findSearchToken({
                text: '(defn skip-ws! []',
                position: { line: 0, character: 6 },
                lineRegexes: [],
                blockCommentStyles: [],
                identCharPattern: /[\w!?-]/,
            }),
            {
                searchToken: 'skip-ws!',
                isString: false,
                isComment: false,
            }
        )
    })

    it('identifies comments after the token', () => {
        assert.deepStrictEqual(
            findSearchToken({
                text: 'foo bar // baz',
                position: { line: 0, character: 5 },
                lineRegexes: [slashPattern],
                blockCommentStyles: [],
            }),
            {
                searchToken: 'bar',
                isString: false,
                isComment: false,
            }
        )
    })

    it('identifies comments before the token', () => {
        assert.deepStrictEqual(
            findSearchToken({
                text: 'foo // bar baz',
                position: { line: 0, character: 8 },
                lineRegexes: [slashPattern],
                blockCommentStyles: [],
            }),
            {
                searchToken: 'bar',
                isString: false,
                isComment: true,
            }
        )
    })

    it('special-cases comment content that looks like a function call', () => {
        assert.deepStrictEqual(
            findSearchToken({
                text: 'foo // bar(baz)',
                position: { line: 0, character: 8 },
                lineRegexes: [slashPattern],
                blockCommentStyles: [],
            }),
            {
                searchToken: 'bar',
                isString: false,
                isComment: false,
            }
        )
    })

    it('special-cases comment content that looks like a field projection', () => {
        assert.deepStrictEqual(
            findSearchToken({
                text: 'foo // .bar baz',
                position: { line: 0, character: 9 },
                lineRegexes: [slashPattern],
                blockCommentStyles: [],
            }),
            {
                searchToken: 'bar',
                isString: false,
                isComment: false,
            }
        )
    })

    it('identifies disjoint block comments after the token', () => {
        assert.deepStrictEqual(
            findSearchToken({
                text: 'foo /* bar baz */',
                position: { line: 0, character: 1 },
                lineRegexes: [],
                blockCommentStyles: [cStyleBlockComment],
            }),
            {
                searchToken: 'foo',
                isString: false,
                isComment: false,
            }
        )
    })

    it('identifies disjoint block comments before the token', () => {
        assert.deepStrictEqual(
            findSearchToken({
                text: '/* foo bar */ baz',
                position: { line: 0, character: 15 },
                lineRegexes: [],
                blockCommentStyles: [cStyleBlockComment],
            }),
            {
                searchToken: 'baz',
                isString: false,
                isComment: false,
            }
        )
    })

    it('identifies block comments on same line', () => {
        assert.deepStrictEqual(
            findSearchToken({
                text: 'foo /* bar */ baz',
                position: { line: 0, character: 8 },
                lineRegexes: [],
                blockCommentStyles: [cStyleBlockComment],
            }),
            {
                searchToken: 'bar',
                isString: false,
                isComment: true,
            }
        )
    })

    it('identifies disjoint block comments over multiple lines', () => {
        assert.deepStrictEqual(
            findSearchToken({
                text: [
                    '/* short comment */',
                    '/* comment on this line',
                    'extend to the next line */',
                    '',
                    'foo bar baz',
                    '',
                    "/* another comment that doesn't close",
                ].join('\n`'),
                position: { line: 4, character: 5 },
                lineRegexes: [],
                blockCommentStyles: [cStyleBlockComment],
            }),
            {
                searchToken: 'bar',
                isString: false,
                isComment: false,
            }
        )

        assert.deepStrictEqual(
            findSearchToken({
                text: [
                    'comment with no opening */',
                    '/* another short comment */',
                    '',
                    'foo bar baz',
                    '',
                    '/* comment on this line',
                    'extend to the next line */',
                ].join('\n`'),
                position: { line: 3, character: 5 },
                lineRegexes: [],
                blockCommentStyles: [cStyleBlockComment],
            }),
            {
                searchToken: 'bar',
                isString: false,
                isComment: false,
            }
        )
    })

    it('identifies block comments over multiple lines', () => {
        assert.deepStrictEqual(
            findSearchToken({
                text: ['/*', 'comment', 'foo bar baz', 'comment', '*/'].join('\n`'),
                position: { line: 2, character: 5 },
                lineRegexes: [],
                blockCommentStyles: [cStyleBlockComment],
            }),
            {
                searchToken: 'bar',
                isString: false,
                isComment: true,
            }
        )
    })

    it('identifies strings around the token', () => {
        assert.deepStrictEqual(
            findSearchToken({
                text: '"foo" + bar + \'baz\'',
                position: { line: 0, character: 9 },
                lineRegexes: [slashPattern],
                blockCommentStyles: [],
            }),
            {
                searchToken: 'bar',
                isString: false,
                isComment: false,
            }
        )
    })

    it('identifies strings contents', () => {
        assert.deepStrictEqual(
            findSearchToken({
                text: 'foo + "bar" + baz',
                position: { line: 0, character: 8 },
                lineRegexes: [slashPattern],
                blockCommentStyles: [],
            }),
            {
                searchToken: 'bar',
                isString: true,
                isComment: false,
            }
        )
    })
})
