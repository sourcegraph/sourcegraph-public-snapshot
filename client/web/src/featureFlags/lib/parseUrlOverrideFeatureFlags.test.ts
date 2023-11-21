import { describe, expect, it } from 'vitest'

import { formatUrlOverrideFeatureFlags as format, parseUrlOverrideFeatureFlags } from './parseUrlOverrideFeatureFlags'

function parse(query: string): Record<string, boolean | null> {
    const result = parseUrlOverrideFeatureFlags(query)

    const obj: Record<string, boolean | null> = {}
    for (const [key, value] of result) {
        obj[key] = value
    }
    return obj
}

describe('parseUrlOverrideFeatureFlags', () => {
    it('parses single key', () => {
        expect(parse('')).toEqual({})
        expect(parse('?feat=~foo')).toEqual({ foo: null })
        expect(parse('?feat=-foo')).toEqual({ foo: false })
    })

    it('parses multiple keys', () => {
        expect(parse('?feat=~foo,~bar')).toEqual({ foo: null, bar: null })

        expect(parse('feat=~foo,bar')).toEqual({
            foo: null,
            bar: true,
        })

        expect(parse('feat=-foo,~bar')).toEqual({
            foo: false,
            bar: null,
        })

        expect(parse('feat=-foo,bar')).toEqual({
            foo: false,
            bar: true,
        })
    })

    it('formats feature flags', () => {
        expect(format(new Map())).toEqual([])
        expect(
            format(
                new Map([
                    ['foo', true],
                    ['bar', false],
                    ['baz', null],
                ])
            )
        ).toEqual(['foo', '-bar', '~baz'])
    })
})
