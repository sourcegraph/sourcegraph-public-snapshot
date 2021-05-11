import classnames from 'classnames'
import RefreshIcon from 'mdi-react/RefreshIcon'
import React, { useContext, useEffect, useMemo, useState } from 'react'
import { useHistory } from 'react-router-dom'
import type { PieChartContent } from 'sourcegraph'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { useDebounce } from '@sourcegraph/wildcard/src'

import { ErrorAlert } from '../../../../../../components/alerts'
import { ChartViewContent } from '../../../../../../views/ChartViewContent/ChartViewContent'
import { InsightsApiContext } from '../../../../../core/backend/api-provider'

import styles from './LangStatsInsightLivePreview.module.scss'
import { DEFAULT_PREVIEW_MOCK } from './live-preview-mock-data'

export interface LangStatsInsightLivePreviewProps {
    /** Custom className for the root element of live preview. */
    className?: string
    /** List of repositories for insights. */
    repository: string
    /** Step value for cut off other small values. */
    threshold: number
    /**
     * Disable prop to disable live preview.
     * Used in a consumer of this component when some required fields
     * for live preview are invalid.
     * */
    disabled?: boolean
}

/**
 * Displays live preview chart for creation UI with latest insights settings
 * from creation UI form.
 * */
export const LangStatsInsightLivePreview: React.FunctionComponent<LangStatsInsightLivePreviewProps> = props => {
    const { repository = '', threshold, disabled = false, className } = props

    const history = useHistory()
    const { getLangStatsInsightContent } = useContext(InsightsApiContext)

    const [loading, setLoading] = useState<boolean>(false)
    const [dataOrError, setDataOrError] = useState<PieChartContent<any> | Error | undefined>()
    // Synthetic deps to trigger dry run for fetching live preview data
    const [lastPreviewVersion, setLastPreviewVersion] = useState(0)

    const liveSettings = useMemo(
        () => ({
            repository: repository.trim(),
            threshold: threshold / 100,
        }),
        [repository, threshold]
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

        getLangStatsInsightContent(liveDebouncedSettings)
            .then(data => !hasRequestCanceled && setDataOrError(data))
            .catch(error => !hasRequestCanceled && setDataOrError(asError(error)))
            .finally(() => !hasRequestCanceled && setLoading(false))

        return () => {
            hasRequestCanceled = true
        }
    }, [disabled, lastPreviewVersion, getLangStatsInsightContent, liveDebouncedSettings])

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
                        content={dataOrError ?? DEFAULT_PREVIEW_MOCK}
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
