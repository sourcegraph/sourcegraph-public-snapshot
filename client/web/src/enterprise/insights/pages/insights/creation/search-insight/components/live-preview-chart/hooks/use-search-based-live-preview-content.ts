import { useContext, useEffect, useState } from 'react'

import { Duration } from 'date-fns'

import { asError } from '@sourcegraph/common'
import { useDebounce } from '@sourcegraph/wildcard'

import { CodeInsightsBackendContext } from '../../../../../../../core/backend/code-insights-backend-context'
import { LineChartContent } from '../../../../../../../core/backend/code-insights-backend-types'
import { SearchBasedInsightSeries } from '../../../../../../../core/types'

export interface Input {
    /** Prop to turn off fetching and reset data for live preview chart. */
    disabled: boolean
    series: SearchBasedInsightSeries[]
    step: Duration
    repositories: string[]
}

export interface Output {
    loading: boolean
    dataOrError: LineChartContent<unknown> | Error | undefined
    update: () => void
}

export function useSearchBasedLivePreviewContent(input: Input): Output {
    const { getSearchInsightContent } = useContext(CodeInsightsBackendContext)

    // Synthetic deps to trigger dry run for fetching live preview data
    const [lastPreviewVersion, setLastPreviewVersion] = useState(0)
    const [loading, setLoading] = useState<boolean>(false)
    const [dataOrError, setDataOrError] = useState<LineChartContent<unknown> | Error | undefined>()

    const debouncedInput = useDebounce(input, 500)

    useEffect(() => {
        let hasRequestCanceled = false
        setLoading(true)
        setDataOrError(undefined)

        if (debouncedInput.disabled) {
            setLoading(false)

            return
        }

        getSearchInsightContent({
            insight: debouncedInput,
        })
            .then(data => !hasRequestCanceled && setDataOrError(data))
            .catch(error => !hasRequestCanceled && setDataOrError(asError(error)))
            .finally(() => !hasRequestCanceled && setLoading(false))

        return () => {
            hasRequestCanceled = true
        }
    }, [lastPreviewVersion, getSearchInsightContent, debouncedInput])

    return {
        loading,
        dataOrError,
        update: () => setLastPreviewVersion(count => count + 1),
    }
}
