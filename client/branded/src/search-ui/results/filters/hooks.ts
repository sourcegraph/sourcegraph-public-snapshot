import { Filter } from '@sourcegraph/shared/src/search/stream'
import { useSyncedWithURLState } from '@sourcegraph/wildcard'

export type URLQueryFilter = Pick<Filter, 'kind' | 'label' | 'value'>

export function serializeURLQueryFilters(filters: URLQueryFilter[]): string | null {
    if (filters.length === 0) {
        return null
    }
    return JSON.stringify(filters.map(f => [f.kind, f.label, f.value]))
}

export function deserializeURLQueryFilters(serialized: string | null): URLQueryFilter[] {
    if (!serialized) {
        return []
    }
    const parsed = JSON.parse(serialized) as [Filter['kind'], string, string][]
    return parsed.map(([kind, label, value]) => ({ kind, label, value }))
}

export function useFilterQuery(): [URLQueryFilter[], (newFilters: URLQueryFilter[]) => void] {
    const [filterQuery, setFilterQuery] = useSyncedWithURLState<URLQueryFilter[], string>({
        urlKey: 'filters',
        serializer: serializeURLQueryFilters,
        deserializer: deserializeURLQueryFilters,
        replace: false,
    })

    return [filterQuery, setFilterQuery]
}
