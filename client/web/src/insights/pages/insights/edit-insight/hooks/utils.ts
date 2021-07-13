import { groupBy } from 'lodash'

import { ConfiguredSubjectOrError, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { Settings } from '../../../../../schema/settings.schema'
import { addInsightToDashboard, removeInsightFromDashboard } from '../../../../core/settings-action/dashboards'
import { addInsightToSettings, removeInsightFromSettings } from '../../../../core/settings-action/insights'
import { Insight, INSIGHTS_DASHBOARDS_SETTINGS_KEY } from '../../../../core/types'
import { SUBJECT_SHARING_LEVELS } from '../../../../core/types/subjects'

interface EditInsightProps extends SettingsCascadeProps<Settings> {
    oldInsight: Insight
    newInsight: Insight
}

enum SettingsOperationType {
    addInsight,
    removeInsight,
    removeInsightFromDashboard,
    addInsightToDashboard,
}

interface RemoveInsight {
    type: SettingsOperationType.removeInsight
    subjectId: string
    insightID: string
}

interface AddInsight {
    type: SettingsOperationType.addInsight
    subjectId: string
    insight: Insight
}

interface RemoveInsightFromDashboard {
    type: SettingsOperationType.removeInsightFromDashboard
    subjectId: string
    insightId: string
    dashboardSettingKey: string
}

interface AddInsightToDashboard {
    type: SettingsOperationType.addInsightToDashboard
    subjectId: string
    insightId: string
    dashboardSettingKey: string
}

type SettingsOperation = AddInsight | RemoveInsight | AddInsightToDashboard | RemoveInsightFromDashboard

/**
 * Apply edit operation over jsonc settings string and return serialized final string
 * with all applied edit operations.
 *
 * @param settings - original jsonc setting content
 * @param operations - list of edit operations
 */
export function applyEditOperations(settings: string, operations: SettingsOperation[]): string {
    let settingsContent: string = settings

    for (const operation of operations) {
        switch (operation.type) {
            case SettingsOperationType.addInsight:
                settingsContent = addInsightToSettings(settingsContent, operation.insight)
                continue
            case SettingsOperationType.removeInsight:
                settingsContent = removeInsightFromSettings({
                    originalSettings: settingsContent,
                    insightID: operation.insightID,
                })
                continue
            case SettingsOperationType.addInsightToDashboard:
                settingsContent = addInsightToDashboard(
                    settingsContent,
                    operation.dashboardSettingKey,
                    operation.insightId
                )
                continue
            case SettingsOperationType.removeInsightFromDashboard:
                settingsContent = removeInsightFromDashboard(
                    settingsContent,
                    operation.dashboardSettingKey,
                    operation.insightId
                )
        }
    }

    return settingsContent
}

/**
 * Returns all operation that have to be applied if insight edit operation has happened.
 *
 * This function should simplify the editing of all insight-related entities in the setting cascade.
 * For instance, let's say we have insight in org level dashboard. Then you make this insight private
 * (change the visibility setting to personal) By this operation, we have to do a few edits
 *
 * 1. Remove an insight configuration from the org level setting
 * 2. Remove an insight's id from the all org-level dashboard configurations
 * 3. Put an insight configuration in personal/user settings.
 *
 * To achieve that we have to have low-level API for setting editing. This function absorbs
 * complexity and logic over insight/dashboard management.
 */
export function getUpdatedSubjectSettings(props: EditInsightProps): Record<string, SettingsOperation[]> {
    const { oldInsight, newInsight, settingsCascade } = props

    if (!settingsCascade.subjects) {
        return {}
    }

    const oldInsightSubject = settingsCascade.subjects.find(
        configuredSubject => configuredSubject.subject.id === oldInsight.visibility
    )

    const newInsightSubject = settingsCascade.subjects.find(
        configuredSubject => configuredSubject.subject.id === newInsight.visibility
    )

    if (!oldInsightSubject || !newInsightSubject) {
        return {}
    }

    const editOperations = [
        ...updateInsightSettings({ oldInsight, newInsight, settingsCascade }),
        ...updateDashboardInsightOwnership({ oldInsight, newInsight, settingsCascade }),
        ...updateInsightIdInDashboardIds({ oldInsight, newInsight, settingsCascade }),
    ]

    // Group all operations by subject id to calculate list of subject for updates
    return groupBy(editOperations, operation => operation.subjectId)
}

export function updateInsightSettings(props: EditInsightProps): SettingsOperation[] {
    const { oldInsight, newInsight } = props

    const removeOldInsightOperation: RemoveInsight = {
        type: SettingsOperationType.removeInsight,
        subjectId: oldInsight.visibility,
        insightID: oldInsight.id,
    }

    const addNewInsightOperation: AddInsight = {
        type: SettingsOperationType.addInsight,
        subjectId: newInsight.visibility,
        insight: newInsight,
    }

    return [removeOldInsightOperation, addNewInsightOperation]
}

export function updateDashboardInsightOwnership(props: EditInsightProps): SettingsOperation[] {
    const { oldInsight, newInsight, settingsCascade } = props

    const hasVisibilityChanged = oldInsight.visibility !== newInsight.visibility

    if (!settingsCascade.subjects || !hasVisibilityChanged) {
        return []
    }

    const oldInsightSubject = settingsCascade.subjects.find(
        configuredSubject => configuredSubject.subject.id === oldInsight.visibility
    )

    const newInsightSubject = settingsCascade.subjects.find(
        configuredSubject => configuredSubject.subject.id === newInsight.visibility
    )

    if (!oldInsightSubject || !newInsightSubject) {
        return []
    }

    const previousSubjectType = oldInsightSubject.subject.__typename
    const nextSubjectType = newInsightSubject.subject.__typename

    const previousShareLevel = SUBJECT_SHARING_LEVELS[previousSubjectType]
    const nextShareLevel = SUBJECT_SHARING_LEVELS[nextSubjectType]

    // This means that we increased shared level - insights are still accessible in dashboards
    // of subject with less shared level. Nothing should be changed in settings
    if (nextShareLevel > previousShareLevel) {
        return []
    }

    // Only organizations have the same shared level. This means we have to remove insights from
    // all dashboards of previous setting subject
    if (nextShareLevel === previousShareLevel) {
        return removeInsightFromAllSubjectDashboards(oldInsight, oldInsightSubject)
    }

    // This means we decreased shared level - made an insight more private that it was before edit
    // operation. We have to remove insight id from all dashboards from all subjects with previous
    // shared level.
    if (nextShareLevel < previousShareLevel) {
        const subjectsToUpdate = [
            // Get all subject from old top share level
            ...settingsCascade.subjects.filter(
                configuredSubject => configuredSubject.subject.__typename === previousSubjectType
            ),
            // Get all subjects with new share level but not the subject that will have insight
            ...settingsCascade.subjects
                .filter(configuredSubject => configuredSubject.subject.__typename === nextSubjectType)
                .filter(configuredSubject => configuredSubject.subject.id !== newInsight.visibility),
        ]

        return subjectsToUpdate.flatMap(configuredSubject =>
            removeInsightFromAllSubjectDashboards(oldInsight, configuredSubject)
        )
    }

    return []
}

function updateInsightIdInDashboardIds(props: EditInsightProps): SettingsOperation[] {
    const { oldInsight, newInsight, settingsCascade } = props
    // Since we use camel cased title as id for the insight is users changed title
    // this means id is also changed and we have to re-create insight with new id.
    const hasIdChanged = oldInsight.id !== newInsight.id

    if (!settingsCascade.subjects || !hasIdChanged) {
        return []
    }

    // Remove old insight id from all dashboards and put the new insight id
    return settingsCascade.subjects.flatMap(configuredSubject => {
        const { settings, subject } = configuredSubject

        if (!settings || isErrorLike(settings)) {
            return []
        }

        const dashboards = settings[INSIGHTS_DASHBOARDS_SETTINGS_KEY] ?? {}

        return Object.keys(dashboards)
            .filter(key => dashboards[key]?.insightIds?.includes(oldInsight.id))
            .flatMap(key => {
                const removeOldInsightId: RemoveInsightFromDashboard = {
                    type: SettingsOperationType.removeInsightFromDashboard,
                    dashboardSettingKey: key,
                    insightId: oldInsight.id,
                    subjectId: subject.id,
                }

                const addNewInsightId: AddInsightToDashboard = {
                    type: SettingsOperationType.addInsightToDashboard,
                    dashboardSettingKey: key,
                    insightId: newInsight.id,
                    subjectId: subject.id,
                }

                return [removeOldInsightId, addNewInsightId]
            })
    })
}

function removeInsightFromAllSubjectDashboards(
    insight: Insight,
    configuredSubject: ConfiguredSubjectOrError<Settings>
): RemoveInsightFromDashboard[] {
    const { subject, settings } = configuredSubject

    if (!settings || isErrorLike(settings)) {
        return []
    }

    const dashboards = settings[INSIGHTS_DASHBOARDS_SETTINGS_KEY] ?? {}

    return Object.keys(dashboards)
        .filter(key => dashboards[key]?.insightIds?.includes(insight.id))
        .map(key => ({
            type: SettingsOperationType.removeInsightFromDashboard,
            dashboardSettingKey: key,
            insightId: insight.id,
            subjectId: subject.id,
        }))
}
