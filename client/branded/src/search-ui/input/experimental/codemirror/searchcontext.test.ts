import { EditorState } from '@codemirror/state'
import { describe, expect, it } from 'vitest'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'

import { parseInputAsQuery } from '../../codemirror/parsedQuery'

import { overrideContextOnPaste } from './searchcontext'

describe('overrideContextOnPaste', () => {
    /**
     * `doc` can contain one or two | to denote the current selection.
     */
    function test(doc: string, insert: string): string {
        let from = doc.indexOf('|')
        let to: number | undefined

        if (from > -1) {
            doc = doc.replace(/\|/, '')

            to = doc.indexOf('|')
            if (to > -1) {
                doc = doc.replace(/\|/, '')
            } else {
                to = undefined
            }
        } else {
            from = doc.length
        }

        const state = EditorState.create({
            doc,
            extensions: [overrideContextOnPaste, parseInputAsQuery({ patternType: SearchPatternType.standard })],
            selection: { anchor: from, head: to },
        })
        return state.update({ ...state.replaceSelection(insert), userEvent: 'input.paste' }).state.sliceDoc()
    }

    it('removes the existing global context: filter if it is considered "empty"', () => {
        expect(test('context:global |', 'context:foo bar')).toStrictEqual('context:foo bar')
        expect(test('context:global|', 'bar context:foo')).toStrictEqual('bar context:foo')
    })

    it('keeps the filter the query contains kewords', () => {
        expect(test('context:global foo |', 'OR context:foo bar')).toStrictEqual(
            'context:global foo OR context:foo bar'
        )
        expect(test('context:global |foo OR bar ', 'context:foo bar')).toStrictEqual(
            'context:global context:foo barfoo OR bar '
        )
    })

    it('keeps the filter if the new value does not contain a context filter', () => {
        expect(test('context:global ', 'foo')).toStrictEqual('context:global foo')
        expect(test('context:global | bar', 'foo')).toStrictEqual('context:global foo bar')
    })

    it('keeps the filter if the current value contains the word "context"', () => {
        expect(test('context ', 'context:foo bar')).toStrictEqual('context context:foo bar')
    })

    it('keeps the filter if the new value contains the word "context"', () => {
        expect(test('context:global ', 'context bar')).toStrictEqual('context:global context bar')
        expect(test('context:global ', 'content:"context:foo" bar')).toStrictEqual(
            'context:global content:"context:foo" bar'
        )
    })

    it('does not remove the filter if the new value somehow "breaks" the context filters', () => {
        expect(test('context:global|', 'context:foo bar')).toStrictEqual('context:globalcontext:foo bar')
        expect(test('|context:global ', 'context:foo bar')).toStrictEqual('context:foo barcontext:global ')
    })
})
