import React, { useContext, useEffect, useState } from 'react'
import { ChartContent, LineChartContent } from 'sourcegraph'

import { asError } from '@sourcegraph/common'
import { useDebounce } from '@sourcegraph/wildcard'

import { LivePreviewContainer } from '../../../../../components/creation-ui-kit/live-preview-container/LivePreviewContainer'
import { getSanitizedRepositories } from '../../../../../components/creation-ui-kit/sanitizers/repositories'
import { CodeInsightsBackendContext } from '../../../../../core/backend/code-insights-backend-context'
import { useDistinctValue } from '../../../../../hooks/use-distinct-value'
import { InsightStep } from '../../search-insight/types'
import { getSanitizedCaptureQuery } from '../utils/capture-group-insight-sanitizer'

export const DEFAULT_MOCK_CHART_CONTENT: LineChartContent<any, string> = {
    chart: 'line' as const,
    data: [
        { x: 1588965700286 - 6 * 24 * 60 * 60 * 1000, a: 20, b: 200 },
        { x: 1588965700286 - 5 * 24 * 60 * 60 * 1000, a: 40, b: 177 },
        { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, a: 110, b: 150 },
        { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, a: 105, b: 165 },
        { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, a: 160, b: 100 },
        { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, a: 184, b: 85 },
        { x: 1588965700286, a: 200, b: 50 },
    ],
    series: [
        {
            dataKey: 'a',
            name: 'Go 1.11',
            stroke: 'var(--oc-indigo-7)',
        },
        {
            dataKey: 'b',
            name: 'Go 1.12',
            stroke: 'var(--oc-orange-7)',
        },
    ],
    xAxis: {
        dataKey: 'x',
        scale: 'time' as const,
        type: 'number',
    },
}

interface CaptureGroupCreationLivePreviewProps {
    disabled: boolean
    repositories: string
    query: string
    stepValue: string
    step: InsightStep
    className?: string
}

export const CaptureGroupCreationLivePreview: React.FunctionComponent<CaptureGroupCreationLivePreviewProps> = props => {
    const { disabled, repositories, query, stepValue, step, className } = props
    const { getCaptureInsightContent } = useContext(CodeInsightsBackendContext)
    const [dataOrError, setDataOrError] = useState<ChartContent | Error | undefined>()

    // Synthetic deps to trigger dry run for fetching live preview data
    const [lastPreviewVersion, setLastPreviewVersion] = useState(0)

    const settings = useDistinctValue({
        disabled,
        query: getSanitizedCaptureQuery(query.trim()),
        repositories: getSanitizedRepositories(repositories),
        step: { [step]: stepValue },
    })

    const debouncedSettings = useDebounce(settings, 500)

    useEffect(() => {
        let hasRequestCanceled = false

        setDataOrError(undefined)

        if (debouncedSettings.disabled) {
            setDataOrError(undefined)
            return
        }

        const { query, repositories, step } = debouncedSettings

        getCaptureInsightContent({ query, repositories, step })
            .then(data => !hasRequestCanceled && setDataOrError(data))
            .catch(error => !hasRequestCanceled && setDataOrError(asError(error)))

        return () => {
            hasRequestCanceled = false
        }
    }, [debouncedSettings, getCaptureInsightContent, lastPreviewVersion])

    return (
        <LivePreviewContainer
            dataOrError={dataOrError}
            loading={!disabled && !dataOrError}
            disabled={disabled}
            defaultMock={DEFAULT_MOCK_CHART_CONTENT}
            mockMessage=" The chart preview will be shown here once you have filled out the repositories and series field"
            className={className}
            chartContentClassName="pt-4"
            onUpdateClick={() => setLastPreviewVersion(version => version + 1)}
        />
    )
}
