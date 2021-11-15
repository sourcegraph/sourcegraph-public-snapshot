import React, { useMemo } from 'react'

import { SubmissionErrors } from '../../../../components/form/hooks/useForm'
import { LangStatsInsight } from '../../../../core/types'
import { SupportedInsightSubject } from '../../../../core/types/subjects'
import { LangStatsInsightCreationContent } from '../../creation/lang-stats/components/lang-stats-insight-creation-content/LangStatsInsightCreationContent'
import { LangStatsCreationFormFields } from '../../creation/lang-stats/types'
import { getSanitizedLangStatsInsight } from '../../creation/lang-stats/utils/insight-sanitizer'

export interface EditLangStatsInsightProps {
    insight: LangStatsInsight
    subjects: SupportedInsightSubject[]
    onSubmit: (insight: LangStatsInsight) => SubmissionErrors | Promise<SubmissionErrors> | void
    onCancel: () => void
}

export const EditLangStatsInsight: React.FunctionComponent<EditLangStatsInsightProps> = props => {
    const { insight, subjects, onSubmit, onCancel } = props

    const insightFormValues = useMemo<LangStatsCreationFormFields>(
        () => ({
            title: insight.title,
            repository: insight.repository,
            threshold: insight.otherThreshold * 100,
            visibility: insight.visibility,
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
            subjects={subjects}
            onSubmit={handleSubmit}
            onCancel={onCancel}
        />
    )
}
