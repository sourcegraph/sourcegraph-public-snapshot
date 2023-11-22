import type { QueryState } from '@sourcegraph/shared/src/search'

import type { EditableDataSeries } from '../../../../components'

export type InsightStep = 'hours' | 'days' | 'weeks' | 'months' | 'years'
export type RepoMode = 'search-query' | 'urls-list'

export interface CreateInsightFormFields {
    /** Code Insight series setting (name of line, line query, color) */
    series: EditableDataSeries[]

    /** Title of code insight */
    title: string

    /** Repositories which to be used to get the info for code insights */
    repositories: string[]

    /**
     * [Experimental] Repositories UI can work in different modes when we have
     * two repo UI fields version of the creation UI. This field controls the
     * current mode
     */
    repoMode: RepoMode

    /**
     * Search-powered query, this is used to gather different repositories though
     * search API instead of having strict list of repo URLs.
     */
    repoQuery: QueryState

    /** Setting for set chart step - how often do we collect data. */
    step: InsightStep

    /** Value for insight step setting */
    stepValue: string

    /** The total number of dashboards on which this insight is referenced. */
    dashboardReferenceCount: number
}
