import { Observable, of, throwError } from 'rxjs'
import { switchMap } from 'rxjs/operators'

import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { isDefined } from '@sourcegraph/shared/src/util/types'

import { Settings, InsightDashboard as InsightDashboardConfiguration } from '../../../../schema/settings.schema'
import { getInsightsDashboards } from '../../hooks/use-dashboards/use-dashboards'
import { getInsightIdsFromSettings } from '../../hooks/use-dashboards/utils'
import { getDeleteInsightEditOperations } from '../../hooks/use-delete-insight/delete-helpers'
import { findInsightById } from '../../hooks/use-insight/use-insight'
import { createSanitizedDashboard } from '../../pages/dashboards/creation/utils/dashboard-sanitizer'
import { getReachableInsights } from '../../pages/dashboards/dashboard-page/components/add-insight-modal/utils/get-reachable-insights'
import { findDashboardByUrlId } from '../../pages/dashboards/dashboard-page/components/dashboards-content/utils/find-dashboard-by-url-id'
import { getUpdatedSubjectSettings } from '../../pages/insights/edit-insight/hooks/use-update-settings-subjects/get-updated-subject-settings'
import { addDashboardToSettings, removeDashboardFromSettings } from '../settings-action/dashboards'
import { addInsight } from '../settings-action/insights'
import { Insight, InsightDashboard, InsightTypePrefix, isRealDashboard } from '../types'
import { isSettingsBasedInsightsDashboard } from '../types/dashboard/real-dashboard'
import { isSubjectInsightSupported, SupportedInsightSubject } from '../types/subjects'

import { getBackendInsight } from './api/get-backend-insight'
import { getBuiltInInsight } from './api/get-built-in-insight'
import { getSubjectSettings, updateSubjectSettings } from './api/subject-settings'
import { CodeInsightsBackend } from './code-insights-backend'
import {
    DashboardCreateInput,
    DashboardDeleteInput,
    DashboardUpdateInput,
    FindInsightByNameInput,
    InsightCreateInput,
    InsightUpdateInput,
    ReachableInsight,
} from './code-insights-backend-types'
import { persistChanges } from './utils/persist-changes'

const errorMockMethod = (methodName: string) => () => throwError(new Error(`Implement ${methodName} method first`))

export class CodeInsightsSettingsCascadeBackend implements CodeInsightsBackend {
    constructor(
        private settingCascade: SettingsCascadeOrError<Settings>,
        private platformContext: Pick<PlatformContext, 'updateSettings'>
    ) {}

    // Insights
    public getInsights = (ids?: string[]): Observable<Insight[]> => {
        if (ids) {
            // Return filtered by ids list of insights
            return of(ids.map(id => findInsightById(this.settingCascade, id)).filter(isDefined))
        }

        // Return all insights
        const { final } = this.settingCascade
        const normalizedFinalSettings = !final || isErrorLike(final) ? {} : final
        const insightIds = getInsightIdsFromSettings(normalizedFinalSettings)

        return of(insightIds.map(id => findInsightById(this.settingCascade, id)).filter(isDefined))
    }

    public getInsightById = errorMockMethod('getInsightById')

    public findInsightByName(input: FindInsightByNameInput): Observable<Insight | null> {
        const { name } = input

        return this.getInsights().pipe(
            switchMap(insights => {
                const possibleInsight = insights.find(insight => insight.title === name)

                return of(possibleInsight ?? null)
            })
        )
    }

    public getReachableInsights = (ownerId: string): Observable<ReachableInsight[]> =>
        of(getReachableInsights({ settingsCascade: this.settingCascade, ownerId }))

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

    public getDashboardById = (dashboardId?: string): Observable<InsightDashboard | null> =>
        this.getDashboards().pipe(
            switchMap(dashboards => of(findDashboardByUrlId(dashboards, dashboardId ?? '') ?? null))
        )

    public findDashboardByName = (name: string): Observable<InsightDashboard | null> =>
        this.getDashboards().pipe(
            switchMap(dashboards => {
                const possibleDashboard = dashboards
                    .filter(isRealDashboard)
                    .filter(isSettingsBasedInsightsDashboard)
                    .find(dashboard => dashboard.title === name)

                return of(possibleDashboard ?? null)
            })
        )

    public createDashboard = (input: DashboardCreateInput): Observable<void> =>
        getSubjectSettings(input.visibility).pipe(
            switchMap(settings => {
                const dashboard = createSanitizedDashboard(input)
                const editedSettings = addDashboardToSettings(settings.contents, dashboard)

                return updateSubjectSettings(this.platformContext, input.visibility, editedSettings)
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

    public updateDashboard = (input: DashboardUpdateInput): Observable<void> => {
        const { previousDashboard, nextDashboardInput } = input

        return of(null).pipe(
            switchMap(() => {
                if (previousDashboard.owner.id !== nextDashboardInput.visibility) {
                    return getSubjectSettings(previousDashboard.owner.id).pipe(
                        switchMap(settings => {
                            const editedSettings = removeDashboardFromSettings(
                                settings.contents,
                                previousDashboard.settingsKey
                            )

                            return updateSubjectSettings(
                                this.platformContext,
                                previousDashboard.owner.id,
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
                    settingsContent = removeDashboardFromSettings(settingsContent, previousDashboard.settingsKey)
                }

                const updatedDashboard: InsightDashboardConfiguration = {
                    ...createSanitizedDashboard(nextDashboardInput),
                    // We have to preserve id and insights IDs value since edit UI
                    // doesn't have these options.
                    id: previousDashboard.id,
                    insightIds: nextDashboardInput.insightIds ?? previousDashboard.insightIds,
                }

                settingsContent = addDashboardToSettings(settingsContent, updatedDashboard)

                return updateSubjectSettings(this.platformContext, nextDashboardInput.visibility, settingsContent)
            })
        )
    }

    // Live preview fetchers
    public getSearchInsightContent = () => errorMockMethod('getSearchInsightContent')().toPromise()
    public getLangStatsInsightContent = () => errorMockMethod('getLangStatsInsightContent')().toPromise()

    // Repositories API
    public getRepositorySuggestions = () => errorMockMethod('getRepositorySuggestions')().toPromise()
    public getResolvedSearchRepositories = () => errorMockMethod('getResolvedSearchRepositories')().toPromise()
}
