import { useContext, useEffect, useState } from 'react'

import { Duration } from 'date-fns'

import { asError } from '@sourcegraph/common'
import { useDebounce } from '@sourcegraph/wildcard'

import { CodeInsightsBackendContext } from '../../../../../../core/backend/code-insights-backend-context'
import { LineChartContent } from '../../../../../../core/backend/code-insights-backend-types'

interface Input {
    disabled: boolean
    query: string
    repositories: string[]
    step: Duration
}

interface Output {
    loading: boolean
    dataOrError: LineChartContent<unknown> | Error | undefined
    update: () => void
}

export function useCaptureGroupPreviewContent(input: Input): Output {
    const { getCaptureInsightContent } = useContext(CodeInsightsBackendContext)

    const [lastPreviewVersion, setLastPreviewVersion] = useState(0)
    const [loading, setLoading] = useState<boolean>(false)
    const [dataOrError, setDataOrError] = useState<LineChartContent<unknown> | Error | undefined>()
    // Synthetic deps to trigger dry run for fetching live preview data
    const debouncedSettings = useDebounce(input, 500)

    useEffect(() => {
        let hasRequestCanceled = false

        setDataOrError(undefined)

        if (debouncedSettings.disabled) {
            setLoading(false)
            return
        }

        const { query, repositories, step } = debouncedSettings

        getCaptureInsightContent({ query, repositories, step })
            .then(data => !hasRequestCanceled && setDataOrError(data))
            .catch(error => !hasRequestCanceled && setDataOrError(asError(error)))
            .finally(() => !hasRequestCanceled && setLoading(false))

        return () => {
            hasRequestCanceled = false
        }
    }, [debouncedSettings, getCaptureInsightContent, lastPreviewVersion])

    return {
        loading,
        dataOrError,
        update: () => setLastPreviewVersion(count => count + 1),
    }
}
