import React, { useContext, useEffect, useMemo, useState } from 'react'
import type { LineChartContent } from 'sourcegraph'

import { asError } from '@sourcegraph/shared/src/util/errors'
import { useDebounce } from '@sourcegraph/wildcard'

import { LivePreviewContainer } from '../../../../../../components/live-preview-container/LivePreviewContainer'
import { InsightsApiContext } from '../../../../../../core/backend/api-provider'
import { SearchBasedInsightSeries } from '../../../../../../core/types/insight/search-insight'
import { useDistinctValue } from '../../../../../../hooks/use-distinct-value'
import { EditableDataSeries, InsightStep } from '../../types'
import { getSanitizedLine, getSanitizedRepositories } from '../../utils/insight-sanitizer'

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

    const { getSearchInsightContent } = useContext(InsightsApiContext)

    const [loading, setLoading] = useState<boolean>(false)
    const [dataOrError, setDataOrError] = useState<LineChartContent<any, string> | Error | undefined>()
    // Synthetic deps to trigger dry run for fetching live preview data
    const [lastPreviewVersion, setLastPreviewVersion] = useState(0)

    const liveSeries = useDistinctValue(
        series
            .filter(series => series.valid)
            // Cut off all unnecessary for live preview fields in order to
            // not trigger live preview update if any of unnecessary has been updated
            // Example: edit true => false - chart shouldn't re-fetch data
            .map<SearchBasedInsightSeries>(getSanitizedLine)
    )

    const liveSettings = useMemo(
        () => ({
            series: liveSeries,
            repositories: getSanitizedRepositories(repositories),
            step: { [step]: stepValue },
        }),
        [step, stepValue, liveSeries, repositories]
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

        getSearchInsightContent(liveDebouncedSettings, { where: 'insightsPage', context: {} })
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
            onUpdateClick={() => setLastPreviewVersion(version => version + 1)}
        />
    )
}
