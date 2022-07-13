import React, { useMemo } from 'react'

import { SubmissionErrors } from '../../../../components'
import { MinimalLangStatsInsightData, LangStatsInsight } from '../../../../core'
import { LangStatsCreationFormFields } from '../../creation/lang-stats'
import { LangStatsInsightCreationContent } from '../../creation/lang-stats/components/LangStatsInsightCreationContent'
import { getSanitizedLangStatsInsight } from '../../creation/lang-stats/utils/insight-sanitizer'

export interface EditLangStatsInsightProps {
    insight: LangStatsInsight
    onSubmit: (insight: MinimalLangStatsInsightData) => SubmissionErrors | Promise<SubmissionErrors> | void
    onCancel: () => void
}

export const EditLangStatsInsight: React.FunctionComponent<
    React.PropsWithChildren<EditLangStatsInsightProps>
> = props => {
    const { insight, onSubmit, onCancel } = props

    const insightFormValues = useMemo<LangStatsCreationFormFields>(
        () => ({
            title: insight.title,
            repository: insight.repository,
            threshold: insight.otherThreshold * 100,
            dashboardReferenceCount: insight.dashboardReferenceCount,
        }),
        [insight]
    )

    // Handlers
    const handleSubmit = (values: LangStatsCreationFormFields): SubmissionErrors | Promise<SubmissionErrors> | void => {
        const sanitizedInsight = getSanitizedLangStatsInsight(values)

        return onSubmit(sanitizedInsight)
    }

    return (
        <LangStatsInsightCreationContent
            mode="edit"
            className="pb-5"
            initialValues={insightFormValues}
            insight={insight}
            onSubmit={handleSubmit}
            onCancel={onCancel}
        />
    )
}
