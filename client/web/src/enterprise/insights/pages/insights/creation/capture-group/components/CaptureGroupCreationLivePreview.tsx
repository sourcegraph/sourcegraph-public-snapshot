import React, { useState } from 'react'
import { LineChartContent } from 'sourcegraph'

import { LivePreviewContainer } from '../../../../../components/live-preview-container/LivePreviewContainer'

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
    className?: string
}

export const CaptureGroupCreationLivePreview: React.FunctionComponent<CaptureGroupCreationLivePreviewProps> = props => {
    const { className } = props

    // Synthetic deps to trigger dry run for fetching live preview data
    const [lastPreviewVersion, setLastPreviewVersion] = useState(0)

    console.log(lastPreviewVersion)

    return (
        <LivePreviewContainer
            dataOrError={undefined}
            loading={false}
            disabled={true}
            defaultMock={DEFAULT_MOCK_CHART_CONTENT}
            mockMessage=" The chart preview will be shown here once you have filled out the repositories and series field"
            className={className}
            chartContentClassName="pt-4"
            onUpdateClick={() => setLastPreviewVersion(version => version + 1)}
        />
    )
}
