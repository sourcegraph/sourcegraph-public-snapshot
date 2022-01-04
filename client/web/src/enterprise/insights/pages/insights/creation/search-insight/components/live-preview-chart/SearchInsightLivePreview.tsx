import React, { useContext, useEffect, useState } from 'react'
import type { LineChartContent } from 'sourcegraph'

import { asError } from '@sourcegraph/common'
import { useDebounce } from '@sourcegraph/wildcard/src'

import { LivePreviewContainer } from '../../../../../../components/creation-ui-kit/live-preview-container/LivePreviewContainer'
import { getSanitizedRepositories } from '../../../../../../components/creation-ui-kit/sanitizers/repositories'
import { CodeInsightsBackendContext } from '../../../../../../core/backend/code-insights-backend-context'
import { SearchBasedInsightSeries } from '../../../../../../core/types/insight/search-insight'
import { useDistinctValue } from '../../../../../../hooks/use-distinct-value'
import { EditableDataSeries, InsightStep } from '../../types'
import { getSanitizedLine } from '../../utils/insight-sanitizer'

import { DEFAULT_MOCK_CHART_CONTENT } from './live-preview-mock-data'

export interface SearchInsightLivePreviewProps {
    /** Custom className for the root element of live preview. */
    className?: string
    /** List of repositories for insights. */
    repositories: string
    /** All Series for live chart. */
    series: EditableDataSeries[]
    /** Step value for chart. */
    stepValue: string

    /**
     * Disable prop to disable live preview.
     * Used in a consumer of this component when some required fields
     * for live preview are invalid.
     */
    disabled?: boolean

    /** Step mode for step value prop. */
    step: InsightStep

    isAllReposMode: boolean
}

/**
 * Displays live preview chart for creation UI with latest insights settings
 * from creation UI form.
 */
export const SearchInsightLivePreview: React.FunctionComponent<SearchInsightLivePreviewProps> = props => {
    const { series, repositories, step, stepValue, disabled = false, isAllReposMode, className } = props

    const { getSearchInsightContent } = useContext(CodeInsightsBackendContext)

    const [loading, setLoading] = useState<boolean>(false)
    const [dataOrError, setDataOrError] = useState<LineChartContent<any, string> | Error | undefined>()
    // Synthetic deps to trigger dry run for fetching live preview data
    const [lastPreviewVersion, setLastPreviewVersion] = useState(0)

    // Compare live insight settings with deep check to avoid unnecessary
    // search insight content fetching
    const liveSettings = useDistinctValue({
        series: series
            .filter(series => series.valid)
            // Cut off all unnecessary for live preview fields in order to
            // not trigger live preview update if any of unnecessary has been updated
            // Example: edit true => false - chart shouldn't re-fetch data
            .map<SearchBasedInsightSeries>(getSanitizedLine),
        repositories: getSanitizedRepositories(repositories),
        step: { [step]: stepValue },
        disabled,
    })

    const liveDebouncedSettings = useDebounce(liveSettings, 500)

    useEffect(() => {
        let hasRequestCanceled = false
        setLoading(true)
        setDataOrError(undefined)

        if (liveDebouncedSettings.disabled) {
            setLoading(false)

            return
        }

        getSearchInsightContent({
            insight: liveDebouncedSettings,
            options: { where: 'insightsPage', context: {} },
        })
            .then(data => !hasRequestCanceled && setDataOrError(data))
            .catch(error => !hasRequestCanceled && setDataOrError(asError(error)))
            .finally(() => !hasRequestCanceled && setLoading(false))

        return () => {
            hasRequestCanceled = true
        }
    }, [lastPreviewVersion, getSearchInsightContent, liveDebouncedSettings])

    return (
        <LivePreviewContainer
            dataOrError={dataOrError}
            loading={loading}
            disabled={disabled}
            defaultMock={DEFAULT_MOCK_CHART_CONTENT}
            mockMessage={
                isAllReposMode ? (
                    <span> Live previews are currently not available for insights running over all repositories. </span>
                ) : (
                    <span>
                        {' '}
                        The chart preview will be shown here once you have filled out the repositories and series
                        fields.
                    </span>
                )
            }
            description={
                isAllReposMode
                    ? 'Previews are only displayed only if you individually list up to 50 repositories.'
                    : null
            }
            className={className}
            chartContentClassName="pt-4"
            onUpdateClick={() => setLastPreviewVersion(version => version + 1)}
        />
    )
}
