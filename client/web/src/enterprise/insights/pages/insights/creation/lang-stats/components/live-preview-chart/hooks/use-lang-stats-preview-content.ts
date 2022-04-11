import { useContext, useEffect, useState } from 'react'

import { asError } from '@sourcegraph/common'
import { useDebounce } from '@sourcegraph/wildcard'

import { CodeInsightsBackendContext } from '../../../../../../../core/backend/code-insights-backend-context'
import { PieChartContent } from '../../../../../../../core/backend/code-insights-backend-types'

export interface UseLangStatsPreviewContentProps {
    /** Prop to turn off fetching and reset data for live preview chart. */
    disabled: boolean
    /** Settings which needed to fetch data for live preview. */
    repository: string
    otherThreshold: number
}

export interface UseLangStatsPreviewContentAPI {
    loading: boolean
    dataOrError: PieChartContent<object> | Error | undefined
    update: () => void
}

/**
 * Unified logic for fetching data for live preview chart of lang stats insight.
 */
export function useLangStatsPreviewContent(input: UseLangStatsPreviewContentProps): UseLangStatsPreviewContentAPI {
    const { getLangStatsInsightContent } = useContext(CodeInsightsBackendContext)

    const [loading, setLoading] = useState<boolean>(false)
    const [dataOrError, setDataOrError] = useState<PieChartContent<any> | Error | undefined>()

    // Synthetic deps to trigger dry run for fetching live preview data
    const [lastPreviewVersion, setLastPreviewVersion] = useState(0)

    const liveDebouncedSettings = useDebounce(input, 500)

    useEffect(() => {
        let hasRequestCanceled = false
        setLoading(true)
        setDataOrError(undefined)

        if (liveDebouncedSettings.disabled) {
            setLoading(false)

            return
        }

        getLangStatsInsightContent({ insight: liveDebouncedSettings })
            .then(data => !hasRequestCanceled && setDataOrError(data))
            .catch(error => !hasRequestCanceled && setDataOrError(asError(error)))
            .finally(() => !hasRequestCanceled && setLoading(false))

        return () => {
            hasRequestCanceled = true
        }
    }, [lastPreviewVersion, getLangStatsInsightContent, liveDebouncedSettings])

    return {
        loading,
        dataOrError,
        update: () => setLastPreviewVersion(count => count + 1),
    }
}
