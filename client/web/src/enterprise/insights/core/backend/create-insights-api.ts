import { camelCase, groupBy } from 'lodash'
import { forkJoin, Observable, of, throwError } from 'rxjs'
import { switchMap } from 'rxjs/operators'

import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { isDefined } from '@sourcegraph/shared/src/util/types'

import { InsightDashboard as InsightDashboardConfiguration, Settings } from '../../../../schema/settings.schema'
import { getInsightsDashboards } from '../../hooks/use-dashboards/use-dashboards'
import { getInsightIdsFromSettings, getSubjectDashboardByID } from '../../hooks/use-dashboards/utils'
import { getDeleteInsightEditOperations } from '../../hooks/use-delete-insight/delete-helpers'
import { findInsightById } from '../../hooks/use-insight/use-insight'
import { createSanitizedDashboard } from '../../pages/dashboards/creation/utils/dashboard-sanitizer'
import { getReachableInsights } from '../../pages/dashboards/dashboard-page/components/add-insight-modal/hooks/get-reachable-insights'
import { findDashboardByUrlId } from '../../pages/dashboards/dashboard-page/components/dashboards-content/utils/find-dashboard-by-url-id'
import {
    addDashboardToSettings,
    addInsightToDashboard,
    removeDashboardFromSettings,
    updateDashboardInsightIds,
} from '../settings-action/dashboards'
import { applyEditOperations, SettingsOperation } from '../settings-action/edits'
import { addInsightToSettings } from '../settings-action/insights'
import {
    Insight,
    InsightDashboard,
    INSIGHTS_ALL_REPOS_SETTINGS_KEY,
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
    DashboardInfo,
    DashboardInput,
    InsightCreateRequest,
    ReachableInsight,
    UpdateDashboardInput,
    UpdateInsightRequest,
} from './types'
import { usePersistEditOperations } from '../../hooks/use-persist-edit-operations'
import { getUpdatedSubjectSettings } from '../../pages/insights/edit-insight/hooks/use-update-settings-subjects/get-updated-subject-settings'

const addInsight = (settings: string, insight: Insight, dashboard: InsightDashboard | null): string => {
    const dashboardSettingKey =
        !isVirtualDashboard(dashboard) && isSettingsBasedInsightsDashboard(dashboard)
            ? dashboard.settingsKey
            : undefined

    const transforms = [
        (settings: string) => addInsightToSettings(settings, insight),
        (settings: string) =>
            dashboardSettingKey ? addInsightToDashboard(settings, dashboardSettingKey, insight.id) : settings,
    ]

    return transforms.reduce((settings, transformer) => transformer(settings), settings)
}

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
    public getDashboards = (): Observable<InsightDashboard[]> => {
        const { subjects, final } = this.settingCascade

        return of(getInsightsDashboards(subjects, final))
    }

    public getDashboardById = (dashboardId?: string): Observable<InsightDashboard | null> =>
        this.getDashboards().pipe(
            switchMap(dashboards => of(findDashboardByUrlId(dashboards, dashboardId ?? '') ?? null))
        )

    public createInsight = (event: InsightCreateRequest): Observable<void> => {
        const { insight, subjectId, dashboard } = event

        return this.getSubjectSettings(subjectId).pipe(
            switchMap(settings => {
                const updatedSettings = addInsight(settings.contents, insight, dashboard)

                return this.updateSubjectSettings(this.platformContext, subjectId, updatedSettings)
            })
        )
    }

    public updateDashboardInsightIds = (options: DashboardInfo): Observable<void> => {
        const { dashboardOwnerId, dashboardSettingKey, insightIds } = options

        return this.getSubjectSettings(dashboardOwnerId).pipe(
            switchMap(settings => {
                const editedSettings = updateDashboardInsightIds(settings.contents, dashboardSettingKey, insightIds)

                return this.updateSubjectSettings(this.platformContext, dashboardOwnerId, editedSettings)
            })
        )
    }

    public deleteDashboard = (dashboardSettingKey: string, dashboardOwnerId: string): Observable<void> =>
        this.getSubjectSettings(dashboardOwnerId).pipe(
            switchMap(settings => {
                const updatedSettings = removeDashboardFromSettings(settings.contents, dashboardSettingKey)

                return this.updateSubjectSettings(this.platformContext, dashboardOwnerId, updatedSettings)
            })
        )

    public getReachableInsights = (ownerId: string): Observable<ReachableInsight[]> =>
        of(getReachableInsights({ settingsCascade: this.settingCascade, ownerId }))

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

    public getInsightById = (insightId: string): Observable<Insight | null> =>
        this.getInsights([insightId]).pipe(
            switchMap(result => {
                const firstMatch = result[0]

                return of(firstMatch ?? null)
            })
        )

    public updateInsight = (event: UpdateInsightRequest): Observable<void[]> => {
        const editOperations = getUpdatedSubjectSettings({
            ...event,
            settingsCascade: this.settingCascade,
        })

        return this.persistChanges(editOperations)
    }

    public updateInsightDrillDownFilters = (
        insight: SearchBackendBasedInsight,
        filters: SearchBasedBackendFilters
    ): Observable<void> =>
        this.getSubjectSettings(insight.visibility).pipe(
            switchMap(settings => {
                const insightWithNewFilters: SearchBackendBasedInsight = { ...insight, filters }
                const editedSettings = addInsightToSettings(settings.contents, insightWithNewFilters)

                return this.updateSubjectSettings(this.platformContext, insight.visibility, editedSettings)
            })
        )

    public createInsightWithNewFilters = (inputs: CreateInsightWithFiltersInputs): Observable<void> => {
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

    public getDashboard = (dashboardId: string): Observable<SettingsBasedInsightDashboard | null> => {
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

    public updateDashboard = (input: UpdateDashboardInput): Observable<void> => {
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

    public findDashboardByName = (name: string): Observable<InsightDashboardConfiguration | null> => {
        if (isErrorLike(this.settingCascade.final) || !this.settingCascade.final) {
            return of(null)
        }

        const dashboards = this.settingCascade.final['insights.dashboards'] ?? {}

        return of(dashboards[camelCase(name)] ?? null)
    }

    public createDashboard = (input: DashboardInput): Observable<void> =>
        this.getSubjectSettings(input.visibility).pipe(
            switchMap(settings => {
                const dashboard = createSanitizedDashboard(input)
                const editedSettings = addDashboardToSettings(settings.contents, dashboard)

                return this.updateSubjectSettings(this.platformContext, input.visibility, editedSettings)
            })
        )

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

        return this.persistChanges(deleteInsightOperations)
    }

    public findInsightByName(title: string, type: InsightTypePrefix): Observable<Insight | null> {
        return of(this.settingCascade).pipe(
            switchMap(settingCascade => {
                const finalSettings = !isErrorLike(settingCascade.final) ? settingCascade.final ?? {} : {}
                const normalizedSettingsKeys = Object.keys(finalSettings)
                const normalizedInsightAllReposKeys = Object.keys(
                    finalSettings?.[INSIGHTS_ALL_REPOS_SETTINGS_KEY] ?? {}
                )

                const existingInsightNames = new Set(
                    [...normalizedSettingsKeys, ...normalizedInsightAllReposKeys]
                        // According to our convention about insights name <insight type>.insight.<insight name>
                        .filter(key => key.startsWith(`${type}`))
                        .map(key => camelCase(key.split(`${type}.`).pop()))
                )
            })
        )
    }

    private persistChanges = (operations: SettingsOperation[]): Observable<void[]> => {
        const subjectsToUpdate = groupBy(operations, operation => operation.subjectId)

        const subjectUpdateRequests = Object.keys(subjectsToUpdate).map(subjectId => {
            const editOperations = subjectsToUpdate[subjectId]

            return this.getSubjectSettings(subjectId).pipe(
                switchMap(settings => {
                    // Modify this jsonc file according to this subject's operations
                    const nextSubjectSettings = applyEditOperations(settings.contents, editOperations)

                    return this.updateSubjectSettings(this.platformContext, subjectId, nextSubjectSettings)
                })
            )
        })

        return forkJoin(subjectUpdateRequests)
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
    public deleteInsight = errorMockMethod('deleteInsight')
    public getDashboardById = errorMockMethod('getDashboardById')
    public createInsight = errorMockMethod('createInsight')
}
