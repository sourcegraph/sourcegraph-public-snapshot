import { InsightVisibility } from '../../../../core/types'

export interface LangStatsCreationFormFields {
    title: string
    repository: string
    threshold: number
    visibility: InsightVisibility

    /**
     * The total number of dashboards on which this insight is referenced.
     */
    dashboardReferenceCount: number
}
