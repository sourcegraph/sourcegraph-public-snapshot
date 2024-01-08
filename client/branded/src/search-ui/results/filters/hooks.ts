import { useSyncedWithURLState } from '@sourcegraph/wildcard'

export function useFilterQuery(): [string, (filtersQuery: string) => void] {
    const [filterQuery, setFilterQuery] = useSyncedWithURLState<string, string>({
        urlKey: 'filtersQuery',
        serializer: query => query,
        deserializer: query => query ?? '',
        replace: false,
    })

    return [filterQuery, setFilterQuery]
}
