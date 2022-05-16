import React, { useContext, useMemo } from 'react'

import { useHistory } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner, useObservable } from '@sourcegraph/wildcard'

import { CodeInsightsBackendContext, CreationInsightInput } from '../../../core'
import { useQueryParameters } from '../../../hooks/use-query-parameters'

import { CaptureGroupCreationPage } from './capture-group'
import { LangStatsInsightCreationPage } from './lang-stats/LangStatsInsightCreationPage'
import { SearchInsightCreationPage } from './search-insight'

export enum InsightCreationPageType {
    LangStats = 'lang-stats',
    Search = 'search-based',
    CaptureGroup = 'capture-group',
}

interface InsightCreateEvent {
    insight: CreationInsightInput
}

interface InsightCreationPageProps extends TelemetryProps {
    mode: InsightCreationPageType
}

export const InsightCreationPage: React.FunctionComponent<
    React.PropsWithChildren<InsightCreationPageProps>
> = props => {
    const { mode, telemetryService } = props

    const history = useHistory()
    const { getDashboardById, createInsight } = useContext(CodeInsightsBackendContext)

    const { dashboardId } = useQueryParameters(['dashboardId'])
    const dashboard = useObservable(useMemo(() => getDashboardById({ dashboardId }), [getDashboardById, dashboardId]))

    if (dashboard === undefined) {
        return <LoadingSpinner inline={false} />
    }

    const handleInsightCreateRequest = async (event: InsightCreateEvent): Promise<unknown> => {
        const { insight } = event

        return createInsight({ insight, dashboard }).toPromise()
    }

    const handleInsightSuccessfulCreation = (): void => {
        if (!dashboard) {
            // Navigate to the dashboard page with new created dashboard
            history.push('/insights/dashboards/')

            return
        }

        history.push(`/insights/dashboards/${dashboard.id}`)
    }

    const handleCancel = (): void => {
        history.push(`/insights/dashboards/${dashboard?.id ?? 'all'}`)
    }

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
                telemetryService={telemetryService}
                onInsightCreateRequest={handleInsightCreateRequest}
                onSuccessfulCreation={handleInsightSuccessfulCreation}
                onCancel={handleCancel}
            />
        )
    }

    return (
        <LangStatsInsightCreationPage
            telemetryService={telemetryService}
            onInsightCreateRequest={handleInsightCreateRequest}
            onSuccessfulCreation={handleInsightSuccessfulCreation}
            onCancel={handleCancel}
        />
    )
}
