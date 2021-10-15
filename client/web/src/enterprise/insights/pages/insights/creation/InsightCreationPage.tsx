import React, { useContext } from 'react'
import { useHistory } from 'react-router'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { InsightsApiContext } from '../../../core/backend/api-provider'
import { addInsight } from '../../../core/settings-action/insights'
import { InsightDashboard, isVirtualDashboard, Insight } from '../../../core/types'
import { isUserSubject } from '../../../core/types/subjects'
import { useDashboard } from '../../../hooks/use-dashboard'
import { useInsightSubjects } from '../../../hooks/use-insight-subjects/use-insight-subjects'
import { useQueryParameters } from '../../../hooks/use-query-parameters'

import { LangStatsInsightCreationPage } from './lang-stats/LangStatsInsightCreationPage'
import { SearchInsightCreationPage } from './search-insight/SearchInsightCreationPage'

export enum InsightCreationPageType {
    LangStats = 'lang-stats',
    Search = 'search-based',
}

const getVisibilityFromDashboard = (dashboard: InsightDashboard | null): string | undefined => {
    if (!dashboard || isVirtualDashboard(dashboard)) {
        return undefined
    }

    return dashboard.owner.id
}

interface InsightCreateEvent {
    subjectId: string
    insight: Insight
}

interface InsightCreationPageProps
    extends PlatformContextProps<'updateSettings'>,
        SettingsCascadeProps,
        TelemetryProps {
    mode: InsightCreationPageType
}

export const InsightCreationPage: React.FunctionComponent<InsightCreationPageProps> = props => {
    const { mode, platformContext, settingsCascade, telemetryService } = props

    const history = useHistory()
    const { getSubjectSettings, updateSubjectSettings } = useContext(InsightsApiContext)
    const { dashboardId } = useQueryParameters(['dashboardId'])
    const dashboard = useDashboard({ settingsCascade, dashboardId })

    const subjects = useInsightSubjects({ settingsCascade })

    // Calculate initial value for the visibility setting
    const personalVisibility = subjects.find(isUserSubject)?.id ?? ''
    const dashboardBasedVisibility = getVisibilityFromDashboard(dashboard)
    const insightVisibility = dashboardBasedVisibility ?? personalVisibility

    const handleInsightCreateRequest = async (event: InsightCreateEvent): Promise<void> => {
        const { insight, subjectId } = event

        const settings = await getSubjectSettings(subjectId).toPromise()
        const updatedSettings = addInsight(settings.contents, insight, dashboard)

        await updateSubjectSettings(platformContext, subjectId, updatedSettings).toPromise()
    }

    const handleInsightSuccessfulCreation = (insight: Insight): void => {
        if (!dashboard || isVirtualDashboard(dashboard)) {
            // Navigate to the dashboard page with new created dashboard
            history.push(`/insights/dashboards/${insight.visibility}`)

            return
        }

        if (dashboard.owner.id === insight.visibility) {
            history.push(`/insights/dashboards/${dashboard.id}`)
        } else {
            history.push(`/insights/dashboards/${insight.visibility}`)
        }
    }

    const handleCancel = (): void => {
        history.push(`/insights/dashboards/${dashboard?.id ?? 'all'}`)
    }

    if (mode === InsightCreationPageType.Search) {
        return (
            <SearchInsightCreationPage
                visibility={insightVisibility}
                settingsCascade={settingsCascade}
                telemetryService={telemetryService}
                onInsightCreateRequest={handleInsightCreateRequest}
                onSuccessfulCreation={handleInsightSuccessfulCreation}
                onCancel={handleCancel}
            />
        )
    }

    return (
        <LangStatsInsightCreationPage
            visibility={insightVisibility}
            settingsCascade={settingsCascade}
            telemetryService={telemetryService}
            onInsightCreateRequest={handleInsightCreateRequest}
            onSuccessfulCreation={handleInsightSuccessfulCreation}
            onCancel={handleCancel}
        />
    )
}
