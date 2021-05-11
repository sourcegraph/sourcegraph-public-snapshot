import classnames from 'classnames'
import RefreshIcon from 'mdi-react/RefreshIcon'
import React, { useContext, useEffect, useMemo, useState } from 'react'
import { useHistory } from 'react-router-dom'
import type { LineChartContent } from 'sourcegraph'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { useDebounce } from '@sourcegraph/wildcard/src'

import { ErrorAlert } from '../../../../../../components/alerts'
import { ChartViewContent } from '../../../../../../views/ChartViewContent/ChartViewContent'
import { InsightsApiContext } from '../../../../../core/backend/api-provider'
import { DataSeries } from '../../../../../core/backend/types'
import { InsightStep } from '../../types'
import { getSanitizedRepositories, getSanitizedSeries } from '../../utils/insight-sanitizer'

import { DEFAULT_MOCK_CHART_CONTENT } from './live-preview-mock-data'
import styles from './SearchInsightLivePreview.module.scss'

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

    const history = useHistory()
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
        <div className={classnames(styles.livePreview, className)}>
            <button
                type="button"
                disabled={disabled}
                className={classnames('btn btn-light', styles.livePreviewUpdateButton)}
                onClick={() => setLastPreviewVersion(version => version + 1)}
            >
                Update live preview
                <RefreshIcon size="1rem" className={styles.livePreviewUpdateButtonIcon} />
            </button>

            {loading && (
                <div
                    className={classnames(
                        styles.livePreviewLoader,
                        'flex-grow-1 d-flex flex-column align-items-center justify-content-center'
                    )}
                >
                    <LoadingSpinner /> Loading code insight
                </div>
            )}

            {isErrorLike(dataOrError) && <ErrorAlert className="m-0" error={dataOrError} />}

            {!loading && !isErrorLike(dataOrError) && (
                <div className={styles.livePreviewChartContainer}>
                    <ChartViewContent
                        className={classnames(styles.livePreviewChart, {
                            [styles.livePreviewChartLoading]: !dataOrError,
                        })}
                        history={history}
                        viewID="search-insight-live-preview"
                        telemetryService={NOOP_TELEMETRY_SERVICE}
                        content={dataOrError ?? DEFAULT_MOCK_CHART_CONTENT}
                    />

                    {!dataOrError && (
                        <p className={styles.livePreviewLoadingChartInfo}>
                            Here you’ll see your insight’s chart preview
                        </p>
                    )}
                </div>
            )}
        </div>
    )
}
