import { describe, expect, test } from 'vitest'

import type { Filter } from './FilterControl'
import { getFilterFromURL, urlSearchParamsForFilteredConnection } from './utils'

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

describe('urlSearchParamsForFilteredConnection', () => {
    test('generate correct URL query string', () => {
        expect(
            urlSearchParamsForFilteredConnection({
                pagination: { first: 20 },
                pageSize: 10,
                query: 'test query',
                filterValues: { status: 'open', type: 'issue' },
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
            }).toString()
        ).toBe('existing=param&query=test+query&first=20&status=open&type=issue')
    })

    test('omit default values', () => {
        expect(
            urlSearchParamsForFilteredConnection({
                pagination: { first: 10 },
                pageSize: 10,
                query: '',
                filterValues: { status: 'all' },
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
            }).toString()
        ).toBe('')
    })

    test('omit first/last only when implicit', () => {
        // Implicit `first`.
        expect(
            urlSearchParamsForFilteredConnection({
                pagination: { first: 10 },
                pageSize: 10,
                search: '',
            }).toString()
        ).toBe('')
        expect(
            urlSearchParamsForFilteredConnection({
                pagination: { first: 10, after: 'A' },
                pageSize: 10,
                search: '',
            }).toString()
        ).toBe('after=A')

        // Implicit `last`.
        expect(
            urlSearchParamsForFilteredConnection({
                pagination: { last: 10, before: 'B' },
                pageSize: 10,
                search: '',
            }).toString()
        ).toBe('before=B')

        // Non-implicit `first`.
        expect(
            urlSearchParamsForFilteredConnection({
                pagination: { first: 10, before: 'B' },
                pageSize: 10,
                search: '',
            }).toString()
        ).toBe('first=10&before=B')

        // Non-implicit `last`.
        expect(
            urlSearchParamsForFilteredConnection({
                pagination: { last: 10 },
                pageSize: 10,
                search: '',
            }).toString()
        ).toBe('last=10')
        expect(
            urlSearchParamsForFilteredConnection({
                pagination: { first: 10, last: 10 },
                pageSize: 10,
                search: '',
            }).toString()
        ).toBe('first=10&last=10')
    })

    test('undefined query clears query in URL', () => {
        expect(
            urlSearchParamsForFilteredConnection({
                query: undefined,
                search: 'query=foo',
            }).toString()
        ).toBe('')
    })

    test('preserves existing search', () => {
        expect(
            urlSearchParamsForFilteredConnection({
                query: 'x',
                search: 'foo=bar',
            }).toString()
        ).toBe('foo=bar&query=x')
    })

    test('handle empty input', () => {
        expect(urlSearchParamsForFilteredConnection({ search: '' }).toString()).toBe('')
    })
})
