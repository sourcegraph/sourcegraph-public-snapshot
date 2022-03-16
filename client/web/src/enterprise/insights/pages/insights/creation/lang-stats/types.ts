export interface LangStatsCreationFormFields {
    title: string
    repository: string
    threshold: number

    /**
     * The total number of dashboards on which this insight is referenced.
     */
    dashboardReferenceCount: number
}
