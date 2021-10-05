import React, { useMemo } from 'react'

import { LivePreviewContainer } from '../../../../../../components/live-preview-container/LivePreviewContainer'

import { useLangStatsPreviewContent } from './hooks/use-lang-stats-preview-content'
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

    const previewSetting = useMemo(
        () => ({
            repository: repository.trim(),
            otherThreshold: threshold / 100,
        }),
        [repository, threshold]
    )

    const { loading, dataOrError, update } = useLangStatsPreviewContent({ disabled, previewSetting })

    return (
        <LivePreviewContainer
            dataOrError={dataOrError}
            disabled={disabled}
            className={className}
            loading={loading}
            defaultMock={DEFAULT_PREVIEW_MOCK}
            onUpdateClick={update}
            mockMessage={
                <span>The chart preview will be shown here once you have filled out the repository field.</span>
            }
        />
    )
}
