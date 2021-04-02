import { FilterType } from './filters'
import { Filter } from './token'
import { appendContextFilter, omitFilter, updateFilter, updateFilters } from './transformer'
import { FilterKind, findFilter } from './validate'

describe('appendContextFilter', () => {
    test('appending context to empty query', () => {
        expect(appendContextFilter('', 'ctx')).toEqual('context:ctx ')
    })

    test('appending context to populated query', () => {
        expect(appendContextFilter('foo', 'ctx')).toEqual('context:ctx foo')
    })

    test('appending when query already contains a context', () => {
        expect(appendContextFilter('context:bar foo', 'ctx')).toEqual('context:bar foo')
    })

    test('appending when query already contains multiple contexts', () => {
        expect(appendContextFilter('(context:bar foo) or (context:bar1 foo1)', 'ctx')).toEqual(
            '(context:bar foo) or (context:bar1 foo1)'
        )
    })
})

describe('omitContextFilter', () => {
    const getGlobalContextFilter = (query: string): Filter => {
        const globalContextFilter = findFilter(query, FilterType.context, FilterKind.Global)
        if (!globalContextFilter) {
            throw new Error('Query does not contain a global context filter')
        }
        return globalContextFilter
    }

    test('omit context filter from the start of the query', () => {
        const query = 'context:foo bar'
        expect(omitFilter(query, getGlobalContextFilter(query))).toEqual('bar')
    })

    test('omit context filter from the end of the query', () => {
        const query = 'bar context:foo'
        expect(omitFilter(query, getGlobalContextFilter(query))).toEqual('bar ')
    })

    test('omit context filter from the middle of the query', () => {
        const query = 'bar context:foo bar1'
        expect(omitFilter(query, getGlobalContextFilter(query))).toEqual('bar  bar1')
    })
})

describe('updateFilter', () => {
    test('append count', () => {
        const query = 'content:"count:200"'
        expect(updateFilter(query, 'count', '5000')).toMatchInlineSnapshot('"content:\\"count:200\\" count:5000"')
    })

    test('update first count', () => {
        const query = '(foo count:5) or (bar count:10)'
        expect(updateFilter(query, 'count', '5000')).toMatchInlineSnapshot('"(foo count:5000) or (bar count:10)"')
    })
})

describe('updateFilters (all filters)', () => {
    test('append count', () => {
        const query = 'content:"count:200"'
        expect(updateFilters(query, 'count', '5000')).toMatchInlineSnapshot('"content:\\"count:200\\" count:5000"')
    })

    test('update all counts count', () => {
        const query = '(foo count:5) or (bar count:10)'
        expect(updateFilters(query, 'count', '5000')).toMatchInlineSnapshot('"(foo count:5000) or (bar count:5000)"')
    })
})
