import { SearchBasedInsightSeries } from '../../../../core'

export type InsightStep = 'hours' | 'days' | 'weeks' | 'months' | 'years'

export interface EditableDataSeries extends SearchBasedInsightSeries {
    valid: boolean
    edit: boolean
}

export interface CreateInsightFormFields {
    /**
     * Code Insight series setting (name of line, line query, color)
     */
    series: EditableDataSeries[]

    /**
     * Title of code insight
     */
    title: string

    /**
     * Repositories which to be used to get the info for code insights
     */
    repositories: string

    /**
     * Setting for set chart step - how often do we collect data.
     */
    step: InsightStep

    /**
     * Value for insight step setting
     */
    stepValue: string

    /**
     * This setting stands for turning on/off all repos mode that means this insight
     * will be run over all repos on BE (BE insight)
     */
    allRepos: boolean

    /**
     * The total number of dashboards on which this insight is referenced.
     */
    dashboardReferenceCount: number
}
