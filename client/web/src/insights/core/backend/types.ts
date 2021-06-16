import { Remote } from 'comlink'
import { Duration } from 'date-fns'
import { Observable } from 'rxjs'
import * as sourcegraph from 'sourcegraph'

import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { ViewProviderResult } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'

import { RepositorySuggestion } from './requests/fetch-repository-suggestions'

export enum ViewInsightProviderSourceType {
    Backend = 'Backend',
    Extension = 'Extension',
}

export interface ViewInsightProviderResult extends ViewProviderResult {
    /** The source of view provider to distinguish between data from extension and data from backend */
    source: ViewInsightProviderSourceType
}

export interface SubjectSettingsResult {
    id: number | null
    contents: string
}

export interface SearchInsightSettings {
    series: DataSeries[]
    step: Duration
    repositories: string[]
}

export interface LangStatsInsightsSettings {
    /** URL of git repository from which statistics will be collected */
    repository: string
    /** The threshold below which a language is counted as part of 'Other' */
    threshold: number
}

export interface DataSeries {
    name: string
    stroke: string
    query: string
}

export interface ApiService {
    getCombinedViews: (
        getExtensionsInsights: () => Observable<ViewProviderResult[]>
    ) => Observable<ViewInsightProviderResult[]>

    getInsightCombinedViews: (
        extensionApi: Promise<Remote<FlatExtensionHostAPI>>
    ) => Observable<ViewInsightProviderResult[]>

    getSubjectSettings: (id: string) => Observable<SubjectSettingsResult>

    updateSubjectSettings: (
        context: Pick<PlatformContext, 'updateSettings'>,
        subjectId: string,
        content: string
    ) => Observable<void>
    getSearchInsightContent: (insight: SearchInsightSettings) => Promise<sourcegraph.LineChartContent<any, string>>
    getLangStatsInsightContent: (insight: LangStatsInsightsSettings) => Promise<sourcegraph.PieChartContent<any>>
    getRepositorySuggestions: (query: string) => Promise<RepositorySuggestion[]>
    getResolvedSearchRepositories: (query: string) => Promise<string[]>
}
