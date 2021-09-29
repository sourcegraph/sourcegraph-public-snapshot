import { Insight } from '../../types'

export enum SettingsOperationType {
    addInsight = 'add-insight',
    removeInsight = 'remove-insight',
    removeInsightFromDashboard = 'remove-insight-from-dashboard',
    addInsightToDashboard = 'add-insight-to-dashboard',
}

export interface RemoveInsight {
    type: SettingsOperationType.removeInsight
    subjectId: string
    insightID: string
}

export interface AddInsight {
    type: SettingsOperationType.addInsight
    subjectId: string
    insight: Insight
}

export interface RemoveInsightFromDashboard {
    type: SettingsOperationType.removeInsightFromDashboard
    subjectId: string
    insightId: string
    dashboardSettingKey: string
}

export interface AddInsightToDashboard {
    type: SettingsOperationType.addInsightToDashboard
    subjectId: string
    insightId: string
    dashboardSettingKey: string
}

export type SettingsOperation = AddInsight | RemoveInsight | AddInsightToDashboard | RemoveInsightFromDashboard
