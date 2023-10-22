import * as assert from 'assert'

import { describe, it } from '@jest/globals'

import { cStyleComment, javaStyleComment, leadingHashPattern, pythonStyleComment } from '../language-specs/comments'
import type { CommentStyle } from '../language-specs/language-spec'

import { findDocstring } from './docstrings'

describe('docstrings', () => {
    it('finds nothing when no comment style is specified', () => {
        assert.deepStrictEqual(
            findDocstring({
                fileText: `
        def foo():
            """docstring"""
            pass
        `,
                definitionLine: 1,
                commentStyles: [],
            }),
            undefined
        )
    })

    it('finds one-line python doc block', () => {
        assert.deepStrictEqual(
            findDocstring({
                fileText: `
        def foo():
            """docstring"""
            pass
        `,
                definitionLine: 1,
                commentStyles: [pythonStyleComment],
            }),
            'docstring'
        )
    })

    it('finds multi-line python doc block', () => {
        assert.deepStrictEqual(
            findDocstring({
                fileText: `
        def foo():
            """docstring1
            docstring2"""
            pass
        `,
                definitionLine: 1,
                commentStyles: [pythonStyleComment],
            }),
            'docstring1\ndocstring2'
        )
    })

    it('finds multi-line python doc block 2', () => {
        assert.deepStrictEqual(
            findDocstring({
                fileText: `
        def foo():
            """docstring1
            docstring2
            """
            pass
        `,
                definitionLine: 1,
                commentStyles: [pythonStyleComment],
            }),
            'docstring1\ndocstring2\n'
        )
    })

    it('finds multi-line python doc block 3', () => {
        assert.deepStrictEqual(
            findDocstring({
                fileText: `
        def foo():
            """
            docstring1
            docstring2
            """
            pass
        `,
                definitionLine: 1,
                commentStyles: [pythonStyleComment],
            }),
            '\ndocstring1\ndocstring2\n'
        )
    })

    it('finds single-line C doc', () => {
        assert.deepStrictEqual(
            findDocstring({
                fileText: `
        // docstring
        const foo;
        `,
                definitionLine: 2,
                commentStyles: [cStyleComment],
            }),
            'docstring'
        )
    })

    it('finds multiline single-line C doc', () => {
        assert.deepStrictEqual(
            findDocstring({
                fileText: `
        // docstring1
        // docstring2
        const foo;
        `,
                definitionLine: 3,
                commentStyles: [cStyleComment],
            }),
            'docstring1\ndocstring2'
        )
    })

    it('finds block C doc 1', () => {
        assert.deepStrictEqual(
            findDocstring({
                fileText: `
        /* docstring1
         * docstring2
         */
        const foo;
        `,
                definitionLine: 4,
                commentStyles: [cStyleComment],
            }),
            'docstring1\ndocstring2\n'
        )
    })

    it('finds block C doc 2', () => {
        assert.deepStrictEqual(
            findDocstring({
                fileText: `
        /* docstring1
         * docstring2 */
        const foo;
        `,
                definitionLine: 3,
                commentStyles: [cStyleComment],
            }),
            'docstring1\ndocstring2 '
        )
    })

    it('finds block C doc 3', () => {
        assert.deepStrictEqual(
            findDocstring({
                fileText: `
        /** docstring1
        *docstring2*/
        const foo;
        `,
                definitionLine: 3,
                commentStyles: [cStyleComment],
            }),
            ' docstring1\ndocstring2'
        )
    })

    it('ignores unrelated code between the docstring and definition line', () => {
        assert.deepStrictEqual(
            findDocstring({
                fileText: `
        /**
         * docstring
         */
        @Annotation
        public void FizzBuzz()
        `,
                definitionLine: 5,
                commentStyles: [javaStyleComment],
            }),
            '\ndocstring\n'
        )
    })

    it('searches multiple comment styles', () => {
        const fileText = `
        mod foo {
            //! Comment below the def

            /// Comment above the def
            pub fn new(value: T) -> Rc<T> {
        }
        `

        const commentStyles: CommentStyle[] = [
            {
                ...cStyleComment,
                docstringIgnore: leadingHashPattern,
            },
            {
                lineRegex: /\/\/\/?!?\s?/,
                docstringIgnore: leadingHashPattern,
                docPlacement: 'below the definition',
            },
        ]

        assert.deepStrictEqual(findDocstring({ fileText, definitionLine: 1, commentStyles }), 'Comment below the def')
        assert.deepStrictEqual(findDocstring({ fileText, definitionLine: 5, commentStyles }), 'Comment above the def')
    })
})
