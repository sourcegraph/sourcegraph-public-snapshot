import { camelCase } from 'lodash'
import { Observable, of } from 'rxjs'
import { map, mapTo, switchMap } from 'rxjs/operators'
import { LineChartContent, PieChartContent } from 'sourcegraph'

import { isErrorLike } from '@sourcegraph/common'
import { ViewContexts } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { isDefined } from '@sourcegraph/shared/src/util/types'

import { Settings, InsightDashboard as InsightDashboardConfiguration } from '../../../../../schema/settings.schema'
import { createSanitizedDashboard } from '../../../pages/dashboards/creation/utils/dashboard-sanitizer'
import { getReachableInsights } from '../../../pages/dashboards/dashboard-page/components/add-insight-modal/utils/get-reachable-insights'
import { findDashboardByUrlId } from '../../../pages/dashboards/dashboard-page/components/dashboards-content/utils/find-dashboard-by-url-id'
import { addDashboardToSettings, removeDashboardFromSettings } from '../../settings-action/dashboards'
import { addInsight } from '../../settings-action/insights'
import { Insight, InsightDashboard, InsightTypePrefix, isRealDashboard } from '../../types'
import { isCustomInsightDashboard } from '../../types/dashboard/real-dashboard'
import { ALL_INSIGHTS_DASHBOARD_ID } from '../../types/dashboard/virtual-dashboard'
import { isSubjectInsightSupported, SupportedInsightSubject } from '../../types/subjects'
import { CodeInsightsBackend } from '../code-insights-backend'
import {
    DashboardCreateInput,
    DashboardCreateResult,
    DashboardDeleteInput,
    DashboardUpdateInput,
    DashboardUpdateResult,
    FindInsightByNameInput,
    GetLangStatsInsightContentInput,
    GetSearchInsightContentInput,
    InsightCreateInput,
    InsightUpdateInput,
    ReachableInsight,
} from '../code-insights-backend-types'
import { getBuiltInInsight } from '../core/api/get-built-in-insight'
import { getLangStatsInsightContent } from '../core/api/get-lang-stats-insight-content'
import { getRepositorySuggestions } from '../core/api/get-repository-suggestions'
import { getResolvedSearchRepositories } from '../core/api/get-resolved-search-repositories'
import { getSearchInsightContent } from '../core/api/get-search-insight-content/get-search-insight-content'
import { getSubjectSettings, updateSubjectSettings } from '../core/api/subject-settings'

import { getBackendInsight } from './gql-handlers/get-backend-insight'
import { getDeleteInsightEditOperations } from './utils/delete-helpers'
import { findInsightById } from './utils/find-insight-by-id'
import { getInsightsDashboards, getInsightIdsFromSettings } from './utils/get-insights-dashboards'
import { getUpdatedSubjectSettings } from './utils/get-updated-subject-settings'
import { persistChanges } from './utils/persist-changes'

export class CodeInsightsSettingsCascadeBackend implements CodeInsightsBackend {
    constructor(
        private settingCascade: SettingsCascadeOrError<Settings>,
        private platformContext: Pick<PlatformContext, 'updateSettings'>
    ) {}

    // Insights
    public getInsights = (input: { dashboardId: string }): Observable<Insight[]> =>
        this.getDashboardById({ dashboardId: input.dashboardId }).pipe(
            switchMap(dashboard => {
                if (dashboard) {
                    const ids = dashboard.insightIds

                    if (ids) {
                        // Return filtered by ids list of insights
                        return of(ids.map(id => findInsightById(this.settingCascade, id)).filter(isDefined))
                    }
                }

                // Return all insights
                const { final } = this.settingCascade
                const normalizedFinalSettings = !final || isErrorLike(final) ? {} : final
                const insightIds = getInsightIdsFromSettings(normalizedFinalSettings)

                return of(insightIds.map(id => findInsightById(this.settingCascade, id)).filter(isDefined))
            })
        )

    public getInsightById = (id: string): Observable<Insight | null> => of(findInsightById(this.settingCascade, id))

    public findInsightByName = (input: FindInsightByNameInput): Observable<Insight | null> => {
        const { name } = input

        // Find insight by name among all insights
        return this.getInsights({ dashboardId: ALL_INSIGHTS_DASHBOARD_ID }).pipe(
            switchMap(insights => {
                const possibleInsight = insights.find(insight => insight.title === name)

                return of(possibleInsight ?? null)
            })
        )
    }

    public getReachableInsights = (input: { subjectId: string }): Observable<ReachableInsight[]> =>
        of(getReachableInsights({ settingsCascade: this.settingCascade, ownerId: input.subjectId }))

    public getBackendInsightData = getBackendInsight
    public getBuiltInInsightData = getBuiltInInsight

    public getInsightSubjects = (): Observable<SupportedInsightSubject[]> => {
        if (!this.settingCascade.subjects) {
            return of([])
        }

        return of(
            this.settingCascade.subjects
                .map(configureSubject => configureSubject.subject)
                .filter<SupportedInsightSubject>(isSubjectInsightSupported)
        )
    }

    public createInsight = (input: InsightCreateInput): Observable<void> => {
        const { insight, dashboard } = input

        return getSubjectSettings(insight.visibility).pipe(
            switchMap(settings => {
                const updatedSettings = addInsight(settings.contents, insight, dashboard)

                return updateSubjectSettings(this.platformContext, insight.visibility, updatedSettings)
            })
        )
    }

    public updateInsight = (input: InsightUpdateInput): Observable<void[]> => {
        const editOperations = getUpdatedSubjectSettings({
            ...input,
            settingsCascade: this.settingCascade,
        })

        return persistChanges(this.platformContext, editOperations)
    }

    public deleteInsight = (insightId: string): Observable<void[]> => {
        // For backward compatibility with old code stats insight api we have to delete
        // this insight in a special way. See link below for more information.
        // https://github.com/sourcegraph/sourcegraph-code-stats-insights/blob/master/src/code-stats-insights.ts#L33
        const isOldCodeStatsInsight = insightId === `${InsightTypePrefix.langStats}.language`

        const keyForSearchInSettings = isOldCodeStatsInsight
            ? // Hardcoded value of id from old version of stats insight extension API
              'codeStatsInsights.query'
            : insightId

        const deleteInsightOperations = getDeleteInsightEditOperations({
            insightId: keyForSearchInSettings,
            settingsCascade: this.settingCascade,
        })

        return persistChanges(this.platformContext, deleteInsightOperations)
    }

    // Dashboards
    public getDashboards = (): Observable<InsightDashboard[]> => {
        const { subjects, final } = this.settingCascade

        return of(getInsightsDashboards(subjects, final))
    }

    public getDashboardById = (input: { dashboardId: string | undefined }): Observable<InsightDashboard | null> =>
        this.getDashboards().pipe(
            switchMap(dashboards => of(findDashboardByUrlId(dashboards, input.dashboardId ?? '') ?? null))
        )

    public findDashboardByName = (name: string): Observable<InsightDashboard | null> =>
        this.getDashboards().pipe(
            switchMap(dashboards => {
                const possibleDashboard = dashboards
                    .filter(isRealDashboard)
                    .filter(isCustomInsightDashboard)
                    .find(dashboard => dashboard.title === name)

                return of(possibleDashboard ?? null)
            })
        )

    public getDashboardSubjects = (): Observable<SupportedInsightSubject[]> => this.getInsightSubjects()

    public createDashboard = (input: DashboardCreateInput): Observable<DashboardCreateResult> =>
        getSubjectSettings(input.visibility).pipe(
            switchMap(settings => {
                const dashboard = createSanitizedDashboard(input)
                const editedSettings = addDashboardToSettings(settings.contents, dashboard)

                return updateSubjectSettings(this.platformContext, input.visibility, editedSettings).pipe(
                    mapTo({ id: camelCase(dashboard.title) })
                )
            })
        )

    public deleteDashboard = (input: DashboardDeleteInput): Observable<void> => {
        const { dashboardOwnerId, dashboardSettingKey } = input

        return getSubjectSettings(dashboardOwnerId).pipe(
            switchMap(settings => {
                const updatedSettings = removeDashboardFromSettings(settings.contents, dashboardSettingKey)

                return updateSubjectSettings(this.platformContext, dashboardOwnerId, updatedSettings)
            })
        )
    }

    public updateDashboard = (input: DashboardUpdateInput): Observable<DashboardUpdateResult> => {
        const { previousDashboard, nextDashboardInput } = input

        if (!previousDashboard.owner || !previousDashboard.settingsKey) {
            throw new Error('TODO: Implement updateDashboard for GraphQL API')
        }

        return of(null).pipe(
            switchMap(() => {
                if (previousDashboard.owner!.id !== nextDashboardInput.visibility) {
                    return getSubjectSettings(previousDashboard.owner!.id).pipe(
                        switchMap(settings => {
                            const editedSettings = removeDashboardFromSettings(
                                settings.contents,
                                previousDashboard.settingsKey!
                            )

                            return updateSubjectSettings(
                                this.platformContext,
                                previousDashboard.owner!.id,
                                editedSettings
                            )
                        })
                    )
                }

                return of(null)
            }),
            switchMap(() => getSubjectSettings(nextDashboardInput.visibility)),
            switchMap(settings => {
                let settingsContent = settings.contents

                // Since id (settings key) of insight dashboard is based on its title
                // if title was changed we need remove old dashboard object from the settings
                // by dashboard's old id
                if (previousDashboard.title !== nextDashboardInput.name) {
                    settingsContent = removeDashboardFromSettings(settingsContent, previousDashboard.settingsKey!)
                }

                const updatedDashboard: InsightDashboardConfiguration = {
                    ...createSanitizedDashboard(nextDashboardInput),
                    // We have to preserve id and insights IDs value since edit UI
                    // doesn't have these options.
                    id: previousDashboard.id,
                    insightIds: nextDashboardInput.insightIds ?? previousDashboard.insightIds,
                }

                settingsContent = addDashboardToSettings(settingsContent, updatedDashboard)

                return updateSubjectSettings(this.platformContext, nextDashboardInput.visibility, settingsContent).pipe(
                    map(() => ({ id: camelCase(updatedDashboard.title) }))
                )
            })
        )
    }

    public assignInsightsToDashboard = (input: DashboardUpdateInput): Observable<unknown> => this.updateDashboard(input)

    // Live preview fetchers
    public getSearchInsightContent = <D extends keyof ViewContexts>(
        input: GetSearchInsightContentInput<D>
    ): Promise<LineChartContent<any, string>> => getSearchInsightContent(input.insight, input.options)

    public getLangStatsInsightContent = <D extends keyof ViewContexts>(
        input: GetLangStatsInsightContentInput<D>
    ): Promise<PieChartContent<any>> => getLangStatsInsightContent(input.insight, input.options)

    public getCaptureInsightContent = (): Promise<LineChartContent<any, string>> =>
        Promise.reject(new Error('Setting based api doesnt support capture group insight'))

    // Repositories API
    public getRepositorySuggestions = getRepositorySuggestions
    public getResolvedSearchRepositories = getResolvedSearchRepositories
}
