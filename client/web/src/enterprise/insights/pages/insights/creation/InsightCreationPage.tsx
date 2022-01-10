import React, { useContext, useMemo } from 'react'
import { useHistory } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { CodeInsightsBackendContext } from '../../../core/backend/code-insights-backend-context'
import { parseDashboardScope } from '../../../core/backend/utils/parse-dashboard-scope'
import { InsightDashboard, isVirtualDashboard, Insight } from '../../../core/types'
import { isUserSubject } from '../../../core/types/subjects'
import { useQueryParameters } from '../../../hooks/use-query-parameters'

import { CaptureGroupCreationPage } from './capture-group/CaptureGroupCreationPage'
import { LangStatsInsightCreationPage } from './lang-stats/LangStatsInsightCreationPage'
import { SearchInsightCreationPage } from './search-insight/SearchInsightCreationPage'

export enum InsightCreationPageType {
    LangStats = 'lang-stats',
    Search = 'search-based',
    CaptureGroup = 'capture-group',
}

const getVisibilityFromDashboard = (dashboard: InsightDashboard | null): string | undefined => {
    if (!dashboard || isVirtualDashboard(dashboard)) {
        return undefined
    }

    // If no owner, this is using the graphql api
    if (!dashboard.owner) {
        return parseDashboardScope(dashboard.grants)
    }

    return dashboard.owner.id
}

interface InsightCreateEvent {
    insight: Insight
}

interface InsightCreationPageProps extends TelemetryProps {
    mode: InsightCreationPageType
}

export const InsightCreationPage: React.FunctionComponent<InsightCreationPageProps> = props => {
    const { mode, telemetryService } = props

    const history = useHistory()
    const { getDashboardById, getInsightSubjects, createInsight } = useContext(CodeInsightsBackendContext)

    const { dashboardId } = useQueryParameters(['dashboardId'])
    const dashboard = useObservable(useMemo(() => getDashboardById({ dashboardId }), [getDashboardById, dashboardId]))
    const subjects = useObservable(useMemo(() => getInsightSubjects(), [getInsightSubjects]))

    if (dashboard === undefined || subjects === undefined) {
        return <LoadingSpinner inline={false} />
    }

    const handleInsightCreateRequest = async (event: InsightCreateEvent): Promise<unknown> => {
        const { insight } = event

        return createInsight({ insight, dashboard }).toPromise()
    }

    const handleInsightSuccessfulCreation = (insight: Insight): void => {
        if (!dashboard || isVirtualDashboard(dashboard)) {
            // Navigate to the dashboard page with new created dashboard
            history.push(`/insights/dashboards/${insight.visibility}`)

            return
        }

        if (!dashboard.owner) {
            history.push(`/insights/dashboards/${dashboard.id}`)
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

    // Calculate initial value for the visibility setting
    const personalVisibility = subjects.find(isUserSubject)?.id ?? ''
    const dashboardBasedVisibility = getVisibilityFromDashboard(dashboard)
    const insightVisibility = dashboardBasedVisibility ?? personalVisibility

    if (mode === InsightCreationPageType.CaptureGroup) {
        return (
            <CaptureGroupCreationPage
                telemetryService={telemetryService}
                onInsightCreateRequest={handleInsightCreateRequest}
                onSuccessfulCreation={handleInsightSuccessfulCreation}
                onCancel={handleCancel}
            />
        )
    }

    if (mode === InsightCreationPageType.Search) {
        return (
            <SearchInsightCreationPage
                visibility={insightVisibility}
                subjects={subjects}
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
            subjects={subjects}
            telemetryService={telemetryService}
            onInsightCreateRequest={handleInsightCreateRequest}
            onSuccessfulCreation={handleInsightSuccessfulCreation}
            onCancel={handleCancel}
        />
    )
}
