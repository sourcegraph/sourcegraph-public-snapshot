import type { QueryState } from '@sourcegraph/shared/src/search'

import type { InsightStep } from '../search-insight'

export interface CaptureGroupFormFields {
    /**
     * Repositories which to be used to get the info for code insights
     */
    repositories: string[]

    repoQuery: QueryState

    repoMode: 'search-query' | 'urls-list'

    /**
     * Query to collect all version like series on BE
     */
    groupSearchQuery: string

    /**
     * Title of code insight
     */
    title: string

    /**
     * Setting for set chart step - how often do we collect data.
     */
    step: InsightStep

    /**
     * Value for insight step setting
     */
    stepValue: string

    /**
     * The total number of dashboards on which this insight is referenced.
     */
    dashboardReferenceCount: number
}
