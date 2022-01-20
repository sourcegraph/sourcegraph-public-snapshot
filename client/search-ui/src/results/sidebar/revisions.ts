import { QueryUpdate } from '@sourcegraph/search'

export enum TabIndex {
    BRANCHES,
    TAGS,
}

export interface RevisionsProps {
    repoName: string
    onFilterClick: (updates: QueryUpdate[]) => void
    query: string
    /**
     * This property is only exposed for storybook tests.
     */
    _initialTab?: TabIndex
}
