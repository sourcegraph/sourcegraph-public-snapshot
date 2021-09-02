import { useContext, useEffect, useState } from 'react'
import { PieChartContent } from 'sourcegraph'

import { asError } from '@sourcegraph/shared/src/util/errors'
import { useDebounce } from '@sourcegraph/wildcard'

import { InsightsApiContext } from '../../../../../../../core/backend/api-provider'

export interface UseLangStatsPreviewContentProps {
    /** Prop to turn off fetching and reset data for live preview chart. */
    disabled: boolean
    /** Settings which needed to fetch data for live preview. */
    previewSetting: {
        repository: string
        otherThreshold: number
    }
}

export interface UseLangStatsPreviewContentAPI {
    loading: boolean
    dataOrError: PieChartContent<any> | Error | undefined
    update: () => void
}

/**
 * Unified logic for fetching data for live preview chart of lang stats insight.
 * */
export function useLangStatsPreviewContent(props: UseLangStatsPreviewContentProps): UseLangStatsPreviewContentAPI {
    const { disabled, previewSetting } = props

    const { getLangStatsInsightContent } = useContext(InsightsApiContext)

    const [loading, setLoading] = useState<boolean>(false)
    const [dataOrError, setDataOrError] = useState<PieChartContent<any> | Error | undefined>()

    // Synthetic deps to trigger dry run for fetching live preview data
    const [lastPreviewVersion, setLastPreviewVersion] = useState(0)

    const liveDebouncedSettings = useDebounce(previewSetting, 500)

    useEffect(() => {
        let hasRequestCanceled = false
        setLoading(true)
        setDataOrError(undefined)

        if (disabled) {
            setLoading(false)

            return
        }

        getLangStatsInsightContent(liveDebouncedSettings, { where: 'insightsPage', context: {} })
            .then(data => !hasRequestCanceled && setDataOrError(data))
            .catch(error => !hasRequestCanceled && setDataOrError(asError(error)))
            .finally(() => !hasRequestCanceled && setLoading(false))

        return () => {
            hasRequestCanceled = true
        }
    }, [disabled, lastPreviewVersion, getLangStatsInsightContent, liveDebouncedSettings])

    return {
        loading,
        dataOrError,
        update: () => setLastPreviewVersion(count => count + 1),
    }
}
