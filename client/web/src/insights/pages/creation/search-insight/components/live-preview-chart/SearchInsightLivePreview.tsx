import React, { useContext, useEffect, useMemo, useState } from 'react'
import type { LineChartContent } from 'sourcegraph'

import { asError } from '@sourcegraph/shared/src/util/errors'
import { useDebounce } from '@sourcegraph/wildcard/src'

import { LivePreviewContainer } from '../../../../../components/live-preview-container/LivePreviewContainer'
import { InsightsApiContext } from '../../../../../core/backend/api-provider'
import { DataSeries } from '../../../../../core/backend/types'
import { InsightStep } from '../../types'
import { getSanitizedRepositories, getSanitizedSeries } from '../../utils/insight-sanitizer'

import { DEFAULT_MOCK_CHART_CONTENT } from './live-preview-mock-data'

export interface SearchInsightLivePreviewProps {
    /** Custom className for the root element of live preview. */
    className?: string
    /** List of repositories for insights. */
    repositories: string
    /** All Series for live chart. */
    series: DataSeries[]
    /** Step value for chart. */
    stepValue: string
    /**
     * Disable prop to disable live preview.
     * Used in a consumer of this component when some required fields
     * for live preview are invalid.
     * */
    disabled?: boolean
    /** Step mode for step value prop. */
    step: InsightStep
}

/**
 * Displays live preview chart for creation UI with latest insights settings
 * from creation UI form.
 * */
export const SearchInsightLivePreview: React.FunctionComponent<SearchInsightLivePreviewProps> = props => {
    const { series, repositories, step, stepValue, disabled = false, className } = props

    const { getSearchInsightContent } = useContext(InsightsApiContext)

    const [loading, setLoading] = useState<boolean>(false)
    const [dataOrError, setDataOrError] = useState<LineChartContent<any, string> | Error | undefined>()
    // Synthetic deps to trigger dry run for fetching live preview data
    const [lastPreviewVersion, setLastPreviewVersion] = useState(0)

    const liveSettings = useMemo(
        () => ({
            series: getSanitizedSeries(series),
            repositories: getSanitizedRepositories(repositories),
            step: { [step]: stepValue },
        }),
        [step, stepValue, series, repositories]
    )

    const liveDebouncedSettings = useDebounce(liveSettings, 500)

    useEffect(() => {
        let hasRequestCanceled = false
        setLoading(true)
        setDataOrError(undefined)

        if (disabled) {
            setLoading(false)

            return
        }

        getSearchInsightContent(liveDebouncedSettings)
            .then(data => !hasRequestCanceled && setDataOrError(data))
            .catch(error => !hasRequestCanceled && setDataOrError(asError(error)))
            .finally(() => !hasRequestCanceled && setLoading(false))

        return () => {
            hasRequestCanceled = true
        }
    }, [disabled, lastPreviewVersion, getSearchInsightContent, liveDebouncedSettings])

    return (
        <LivePreviewContainer
            dataOrError={dataOrError}
            loading={loading}
            disabled={disabled}
            defaultMock={DEFAULT_MOCK_CHART_CONTENT}
            mockMessage={
                <span>
                    {' '}
                    Here you’ll see your insight’s chart preview. <br />
                    You need to fill in the repositories and series fields.
                </span>
            }
            className={className}
            onUpdateClick={() => setLastPreviewVersion(version => version + 1)}
        />
    )
}
