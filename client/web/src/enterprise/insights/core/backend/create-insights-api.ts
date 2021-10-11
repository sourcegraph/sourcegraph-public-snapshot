import { camelCase } from 'lodash'
import { Observable, of, throwError } from 'rxjs'
import { switchMap } from 'rxjs/operators'

import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { isDefined } from '@sourcegraph/shared/src/util/types'

import { InsightDashboard as InsightDashboardConfiguration, Settings } from '../../../../schema/settings.schema'
import { getInsightsDashboards } from '../../hooks/use-dashboards/use-dashboards'
import { getSubjectDashboardByID } from '../../hooks/use-dashboards/utils'
import { findInsightById } from '../../hooks/use-insight/use-insight'
import { createSanitizedDashboard } from '../../pages/dashboards/creation/utils/dashboard-sanitizer';
import { getReachableInsights } from '../../pages/dashboards/dashboard-page/components/add-insight-modal/hooks/get-reachable-insights'
import {
    addDashboardToSettings,
    addInsightToDashboard,
    removeDashboardFromSettings,
    updateDashboardInsightIds,
} from '../settings-action/dashboards'
import { addInsightToSettings } from '../settings-action/insights'
import {
    Insight, InsightDashboard,
    InsightTypePrefix,
    isVirtualDashboard,
    SettingsBasedInsightDashboard,
} from '../types'
import { isSettingsBasedInsightsDashboard } from '../types/dashboard/real-dashboard'
import { SearchBackendBasedInsight, SearchBasedBackendFilters } from '../types/insight/search-insight'
import { isSubjectInsightSupported, SupportedInsightSubject } from '../types/subjects'

import { getBackendInsight } from './api/get-backend-insight'
import { getBuiltInInsight } from './api/get-built-in-insight'
import { getLangStatsInsightContent } from './api/get-lang-stats-insight-content'
import { getRepositorySuggestions } from './api/get-repository-suggestions'
import { getResolvedSearchRepositories } from './api/get-resolved-search-repositories'
import { getSearchInsightContent } from './api/get-search-insight-content/get-search-insight-content'
import { getSubjectSettings, updateSubjectSettings } from './api/subject-settings'
import {
    CodeInsightsBackend,
    CreateInsightWithFiltersInputs,
    DashboardInfo, DashboardInput,
    ReachableInsight,
    UpdateDashboardInput,
} from './types'

export class CodeInsightsSettingBasedBackend implements CodeInsightsBackend {
    constructor(private settingCascade: SettingsCascadeOrError<Settings>, private platformContext: PlatformContext) {}

    // Insights loading
    public getBackendInsight = getBackendInsight
    public getBuiltInInsight = getBuiltInInsight

    // Subject operations TODO [VK] remove this setting oriented methods
    public getSubjectSettings = getSubjectSettings
    public updateSubjectSettings = updateSubjectSettings

    // Live preview fetchers
    public getSearchInsightContent = getSearchInsightContent
    public getLangStatsInsightContent = getLangStatsInsightContent

    // Repositories API
    public getRepositorySuggestions = getRepositorySuggestions
    public getResolvedSearchRepositories = getResolvedSearchRepositories

    // NEW API

    public getDashboards(): Observable<InsightDashboard[]> {
        const { subjects, final } = this.settingCascade

        return of(getInsightsDashboards(subjects, final))
    }

    public updateDashboardInsightIds(options: DashboardInfo): Observable<void> {
        const { dashboardOwnerId, dashboardSettingKey, insightIds } = options

        return this.getSubjectSettings(dashboardOwnerId).pipe(
            switchMap(settings => {
                const editedSettings = updateDashboardInsightIds(settings.contents, dashboardSettingKey, insightIds)

                return this.updateSubjectSettings(this.platformContext, dashboardOwnerId, editedSettings)
            })
        )
    }

    public deleteDashboard(dashboardSettingKey: string, dashboardOwnerId: string): Observable<void> {
        return this.getSubjectSettings(dashboardOwnerId).pipe(
            switchMap(settings => {
                const updatedSettings = removeDashboardFromSettings(settings.contents, dashboardSettingKey)

                return updateSubjectSettings(this.platformContext, dashboardOwnerId, updatedSettings)
            })
        )
    }

    public getReachableInsights(ownerId: string): Observable<ReachableInsight[]> {
        return of(getReachableInsights({ settingsCascade: this.settingCascade, ownerId }))
    }

    public getInsights(ids: string[]): Observable<Insight[]> {
        return of(ids.map(id => findInsightById(this.settingCascade, id)).filter(isDefined))
    }

    public updateInsightDrillDownFilters(
        insight: SearchBackendBasedInsight,
        filters: SearchBasedBackendFilters
    ): Observable<void> {
        return this.getSubjectSettings(insight.visibility).pipe(
            switchMap(settings => {
                const insightWithNewFilters: SearchBackendBasedInsight = { ...insight, filters }
                const editedSettings = addInsightToSettings(settings.contents, insightWithNewFilters)

                return updateSubjectSettings(this.platformContext, insight.visibility, editedSettings)
            })
        )
    }

    public createInsightWithNewFilters(inputs: CreateInsightWithFiltersInputs): Observable<void> {
        const { dashboard, insightName, originalInsight, filters } = inputs

        // Get id of insight setting subject (owner of it insight)
        const subjectId = isVirtualDashboard(dashboard) ? originalInsight.visibility : dashboard.owner.id

        // Create new insight by name and last valid filters value
        const newInsight: SearchBackendBasedInsight = {
            ...originalInsight,
            id: `${InsightTypePrefix.search}.${camelCase(insightName)}`,
            title: insightName,
            filters,
        }

        return this.getSubjectSettings(subjectId).pipe(
            switchMap(settings => {
                const updatedSettings = [
                    (settings: string) => addInsightToSettings(settings, newInsight),
                    (settings: string) => {
                        // Virtual and built-in dashboards calculate their insight dynamically in runtime
                        // no need to store insight list for them explicitly
                        if (isVirtualDashboard(dashboard) || !isSettingsBasedInsightsDashboard(dashboard)) {
                            return settings
                        }

                        return addInsightToDashboard(settings, dashboard.settingsKey, newInsight.id)
                    },
                ].reduce((settings, transformer) => transformer(settings), settings.contents)

                return this.updateSubjectSettings(this.platformContext, subjectId, updatedSettings)
            })
        )
    }

    public getInsightSubjects(): Observable<SupportedInsightSubject[]> {
        if (!this.settingCascade.subjects) {
            return of([])
        }

        return of(
            this.settingCascade.subjects
                .map(configureSubject => configureSubject.subject)
                .filter<SupportedInsightSubject>(isSubjectInsightSupported)
        )
    }

    public getDashboard(dashboardId: string): Observable<SettingsBasedInsightDashboard | null> {
        const subjects = this.settingCascade.subjects
        const configureSubject = subjects?.find(
            ({ settings }) => settings && !isErrorLike(settings) && !!settings['insights.dashboards']?.[dashboardId]
        )

        if (!configureSubject || !configureSubject.settings || isErrorLike(configureSubject.settings)) {
            return of(null)
        }

        const { subject, settings } = configureSubject

        return of(getSubjectDashboardByID(subject, settings, dashboardId))
    }

    public updateDashboard(input: UpdateDashboardInput): Observable<void> {
        const { previousDashboard, nextDashboardInput } = input

        return of(null).pipe(
            switchMap(() => {
                if (previousDashboard.owner.id !== nextDashboardInput.visibility) {
                    return this.getSubjectSettings(previousDashboard.owner.id).pipe(
                        switchMap(settings => {
                            const editedSettings = removeDashboardFromSettings(
                                settings.contents,
                                previousDashboard.settingsKey
                            )

                            return this.updateSubjectSettings(
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
                    insightIds: previousDashboard.insightIds,
                }

                settingsContent = addDashboardToSettings(settingsContent, updatedDashboard)

                return this.updateSubjectSettings(this.platformContext, nextDashboardInput.visibility, settingsContent)
            })
        )
    }

    public findDashboardByName(name: string): Observable<InsightDashboardConfiguration | null> {
        if (isErrorLike(this.settingCascade.final) || !this.settingCascade.final) {
            return of(null)
        }

        const dashboards = this.settingCascade.final['insights.dashboards'] ?? {}

        return of(dashboards[camelCase(name)] ?? null)
    }

    public createDashboard(input: DashboardInput): Observable<void> {
        return this.getSubjectSettings(input.visibility).pipe(
            switchMap(settings => {
                const dashboard = createSanitizedDashboard(input)
                const editedSettings = addDashboardToSettings(settings.contents, dashboard)

                return this.updateSubjectSettings(this.platformContext, input.visibility, editedSettings)
            })
        )
    }
}

const errorMockMethod = (methodName: string) => () => throwError(new Error(`Implement ${methodName} method first`))

export class CodeInsightsFakeBackend implements CodeInsightsBackend {
    // Insights loading
    public getBackendInsight = errorMockMethod('getBackendInsight')
    public getBuiltInInsight = errorMockMethod('getBuiltInInsight')
    public getSubjectSettings = errorMockMethod('getSubjectSettings')
    public updateSubjectSettings = errorMockMethod('updateSubjectSettings')

    // Live preview fetchers
    public getSearchInsightContent = () => errorMockMethod('getSearchInsightContent')().toPromise()
    public getLangStatsInsightContent = () => errorMockMethod('getLangStatsInsightContent')().toPromise()

    // Repositories API
    public getRepositorySuggestions = () => errorMockMethod('getRepositorySuggestions')().toPromise()
    public getResolvedSearchRepositories = () => errorMockMethod('getResolvedSearchRepositories')().toPromise()

    // New high level API
    public getDashboards = errorMockMethod('getDashboards')
    public updateDashboardInsightIds = errorMockMethod('updateDashboardInsightIds')
    public deleteDashboard = errorMockMethod('deleteDashboard')
    public getReachableInsights = errorMockMethod('getReachableInsights')
    public getInsights = errorMockMethod('getInsights')
    public updateInsightDrillDownFilters = errorMockMethod('updateInsightDrillDownFilters')
    public createInsightWithNewFilters = errorMockMethod('createInsightWithNewFilters')
    public getInsightSubjects = errorMockMethod('getInsightSubjects')
    public getDashboard = errorMockMethod('getDashboard')
    public updateDashboard = errorMockMethod('updateDashboard')
    public findDashboardByName = errorMockMethod('findDashboardByName')
    public createDashboard = errorMockMethod('createDashboard')
}
