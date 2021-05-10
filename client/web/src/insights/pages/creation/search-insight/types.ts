
/** Creation form fields. */
export interface CreateInsightFormFields {
    /** Code Insight series setting (name of line, line query, color) */
    series: DataSeries[]
    /** Title of code insight*/
    title: string
    /** Repositories which to be used to get the info for code insights */
    repositories: string
    /** Visibility setting which responsible for where insight will appear. */
    visibility: 'personal' | 'organization'
    /** Setting for set chart step - how often do we collect data. */
    step: 'hours' | 'days' | 'weeks' | 'months' | 'years'
    /** Value for insight step setting */
    stepValue: string
}

export interface DataSeries {
    name: string
    color: string
    query: string
}
