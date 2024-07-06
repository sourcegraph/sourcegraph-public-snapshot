import { describe, expect, test } from 'vitest'

import type { Filter } from './FilterControl'
import { getFilterFromURL, getUrlQuery } from './utils'

describe('getFilterFromURL', () => {
    test('correct filter values from URL parameters', () => {
        const searchParams = new URLSearchParams('filter1=value2&filter2=value1')
        const filters: Filter[] = [
            {
                id: 'filter1',
                label: 'Filter 1',
                type: 'select',
                options: [
                    { label: 'Option 1', value: 'value1', args: {} },
                    { label: 'Option 2', value: 'value2', args: {} },
                ],
            },
            {
                id: 'filter2',
                label: 'Filter 2',
                type: 'select',
                options: [
                    { label: 'Option 1', value: 'value1', args: {} },
                    { label: 'Option 2', value: 'value2', args: {} },
                ],
            },
        ]
        expect(getFilterFromURL(searchParams, filters)).toEqual({
            filter1: 'value2',
            filter2: 'value1',
        })
    })

    test('use first option value when URL parameters are not present', () => {
        const searchParams = new URLSearchParams()
        const filters: Filter[] = [
            {
                id: 'filter1',
                label: 'Filter 1',
                type: 'select',
                options: [
                    { label: 'Option 1', value: 'value1', args: {} },
                    { label: 'Option 2', value: 'value2', args: {} },
                ],
            },
        ]
        expect(getFilterFromURL(searchParams, filters)).toEqual({
            filter1: 'value1',
        })
    })

    test('return an empty object when filters are undefined', () => {
        const searchParams = new URLSearchParams('filter1=value1')
        expect(getFilterFromURL(searchParams, undefined)).toEqual({})
    })
})

describe('getUrlQuery', () => {
    test('generate correct URL query string', () => {
        expect(
            getUrlQuery({
                first: { actual: 20, default: 10 },
                query: 'test query',
                filterValues: { status: 'open', type: 'issue' },
                visibleResultCount: 30,
                filters: [
                    {
                        id: 'status',
                        type: 'select',
                        label: 'l',
                        options: [
                            { value: 'all', label: 'l', args: {} },
                            { value: 'open', label: 'l', args: {} },
                        ],
                    },
                    {
                        id: 'type',
                        type: 'select',
                        label: 'l',
                        options: [
                            { value: 'all', label: 'l', args: {} },
                            { value: 'issue', label: 'l', args: {} },
                        ],
                    },
                ],
                search: '?existing=param',
            })
        ).toBe('existing=param&query=test+query&first=20&status=open&type=issue&visible=30')
    })

    test('omit default values', () => {
        expect(
            getUrlQuery({
                first: { actual: 10, default: 10 },
                query: '',
                filterValues: { status: 'all' },
                visibleResultCount: 10,
                filters: [
                    {
                        id: 'status',
                        type: 'select',
                        label: 'l',
                        options: [
                            { value: 'all', label: 'l', args: {} },
                            { value: 'open', label: 'l', args: {} },
                        ],
                    },
                ],
                search: '',
            })
        ).toBe('')
    })

    test('undefined query clears query in URL', () => {
        expect(
            getUrlQuery({
                query: undefined,
                search: 'query=foo',
            })
        ).toBe('')
    })

    test('handle empty input', () => {
        expect(getUrlQuery({ search: '' })).toBe('')
    })
})
