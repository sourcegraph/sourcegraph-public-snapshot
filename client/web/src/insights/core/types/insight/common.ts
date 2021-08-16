export enum InsightTypePrefix {
    search = 'searchInsights.insight',
    langStats = 'codeStatsInsights.insight',
}

export enum InsightType {
    /**
     * This type of insights work via extension API and all their settings
     * (insights configs) stores at the top level of setting subject
     * At the moment we have two insight specific extensions
     *
     * Search based - https://github.com/sourcegraph/sourcegraph-search-insights
     * Lang Stats - https://github.com/sourcegraph/sourcegraph-code-stats-insights
     */
    Extension = 'extension',

    /**
     * This type of insights work via our backend and their settings you can find
     * under special key in our settings subjects (insights.allrepos: {..configs})
     */
    Backend = 'backend',
}

/**
 * These fields are needed only for the code insight FE logic and they are not stored
 * in any settings (insight configurations) fields.
 */
export interface SyntheticInsightFields {
    /**
     * ID of insight <type of insight>.insight.<name of insight>
     */
    id: string

    /**
     * Visibility of insight. Personal, organization or global setting cascade subject.
     */
    visibility: InsightVisibility
}

/**
 * Visibility setting which responsible for where insight will appear.
 * possible value '<user subject id>' | '<org id 1> ... | ... <org id N> | <global subject id>'
 */
export type InsightVisibility = string
