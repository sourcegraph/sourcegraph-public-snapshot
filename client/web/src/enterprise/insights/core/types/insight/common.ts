/**
 * In setting-based api we store all insight configuration in setting cascade with
 * camel cased insight title as an id. In order to make a difference between
 * search and lang stats insights we have a naming convention for these ids
 * id = <type>.<camelCasedTitle>.
 *
 * This type is used only in setting based api.
 * TODO: Remove this when setting-cascade api is deprecated
 */
export enum InsightTypePrefix {
    search = 'searchInsights.insight',
    langStats = 'codeStatsInsights.insight',
}

export enum InsightExecutionType {
    /** This type of insights run on FE via search API. */
    Runtime = 'runtime',

    /**
     * This type of insights work via our backend and and gql API returns these insight with
     * pre-calculated data points.
     */
    Backend = 'backend',
}

export enum InsightType {
    SearchBased = 'SearchBased',
    LangStats = 'LangStats',
    CaptureGroup = 'CaptureGroup',
}

/**
 * These fields are needed only for the code insight FE runtime logic, and they are not stored
 * in any settings (insight configurations or BE) fields.
 */
export interface SyntheticInsightFields {
    id: string
    viewType: InsightType
    visibility: InsightVisibility
    dashboardReferenceCount: number
}

/**
 * Visibility setting which responsible for where insight will appear.
 * possible value '<user subject id>' | '<org id 1> ... | ... <org id N> | <global subject id>'
 */
export type InsightVisibility = string
