import { FC, useContext } from 'react'

import { useNavigate } from 'react-router-dom'

import { useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { CodeInsightsBackendContext, CreationInsightInput } from '../../../core'
import { useQueryParameters } from '../../../hooks'
import { encodeDashboardIdQueryParam } from '../../../routers.constant'

import { CaptureGroupCreationPage } from './capture-group'
import { ComputeInsightCreationPage } from './compute'
import { LangStatsInsightCreationPage } from './lang-stats'
import { SearchInsightCreationPage } from './search-insight'

export enum InsightCreationPageType {
    LangStats = 'lang-stats',
    Compute = 'compute',
    Search = 'search-based',
    CaptureGroup = 'capture-group',
}

interface InsightCreateEvent {
    insight: CreationInsightInput
}

interface InsightCreationPageProps extends TelemetryProps {
    mode: InsightCreationPageType
    isSourcegraphApp: boolean
}

export const InsightCreationPage: FC<InsightCreationPageProps> = props => {
    const { mode, telemetryService, isSourcegraphApp } = props

    const navigate = useNavigate()
    const { createInsight } = useContext(CodeInsightsBackendContext)
    const { dashboardId } = useQueryParameters(['dashboardId'])

    const codeInsightsCompute = useExperimentalFeatures(features => features.codeInsightsCompute)

    const handleInsightCreateRequest = async (event: InsightCreateEvent): Promise<unknown> => {
        const { insight } = event

        return createInsight({ insight, dashboardId: dashboardId ?? null }).toPromise()
    }

    const handleInsightSuccessfulCreation = (): void => {
        if (!dashboardId) {
            // Navigate to the dashboard page with new created dashboard
            navigate('/insights/all')

            return
        }

        navigate(`/insights/dashboards/${dashboardId}`)
    }

    const handleCancel = (): void => {
        if (!dashboardId) {
            navigate('/insights/all')

            return
        }

        navigate(`/insights/dashboards/${dashboardId}`)
    }

    const backCreateUrl = encodeDashboardIdQueryParam('/insights/create', dashboardId)

    if (mode === InsightCreationPageType.CaptureGroup) {
        return (
            <CaptureGroupCreationPage
                backUrl={backCreateUrl}
                telemetryService={telemetryService}
                onInsightCreateRequest={handleInsightCreateRequest}
                onSuccessfulCreation={handleInsightSuccessfulCreation}
                onCancel={handleCancel}
                isSourcegraphApp={isSourcegraphApp}
            />
        )
    }

    if (mode === InsightCreationPageType.Search) {
        return (
            <SearchInsightCreationPage
                backUrl={backCreateUrl}
                telemetryService={telemetryService}
                onInsightCreateRequest={handleInsightCreateRequest}
                onSuccessfulCreation={handleInsightSuccessfulCreation}
                onCancel={handleCancel}
                isSourcegraphApp={isSourcegraphApp}
            />
        )
    }

    if (codeInsightsCompute && mode === InsightCreationPageType.Compute) {
        return (
            <ComputeInsightCreationPage
                backUrl={backCreateUrl}
                telemetryService={telemetryService}
                onInsightCreateRequest={handleInsightCreateRequest}
                onSuccessfulCreation={handleInsightSuccessfulCreation}
                onCancel={handleCancel}
                isSourcegraphApp={isSourcegraphApp}
            />
        )
    }

    return (
        <LangStatsInsightCreationPage
            backUrl={backCreateUrl}
            telemetryService={telemetryService}
            onInsightCreateRequest={handleInsightCreateRequest}
            onSuccessfulCreation={handleInsightSuccessfulCreation}
            onCancel={handleCancel}
            isSourcegraphApp={isSourcegraphApp}
        />
    )
}
