import type { Filter } from '@sourcegraph/shared/src/search/stream'

export type SectionItem = Omit<Filter, 'count'> & {
    count?: Filter['count']
    selected: boolean
}
