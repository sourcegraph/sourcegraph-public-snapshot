import { type FC, useMemo } from 'react'

import { FORM_ERROR, type SubmissionErrors } from '@sourcegraph/wildcard'

import { CodeInsightCreationMode, CodeInsightsCreationActions } from '../../../../components'
import type { MinimalLangStatsInsightData, LangStatsInsight } from '../../../../core'
import type { LangStatsCreationFormFields } from '../../creation/lang-stats'
import { LangStatsInsightCreationContent } from '../../creation/lang-stats/components/LangStatsInsightCreationContent'
import { getSanitizedLangStatsInsight } from '../../creation/lang-stats/utils/insight-sanitizer'

export interface EditLangStatsInsightProps {
    insight: LangStatsInsight
    licensed: boolean
    isEditAvailable: boolean | undefined
    onSubmit: (insight: MinimalLangStatsInsightData) => SubmissionErrors | Promise<SubmissionErrors> | void
    onCancel: () => void
}

export const EditLangStatsInsight: FC<EditLangStatsInsightProps> = props => {
    const { insight, licensed, isEditAvailable, onSubmit, onCancel } = props

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
            initialValues={insightFormValues}
            touched={true}
            className="pb-5"
            onSubmit={handleSubmit}
        >
            {form => (
                <CodeInsightsCreationActions
                    mode={CodeInsightCreationMode.Edit}
                    licensed={licensed}
                    available={isEditAvailable}
                    submitting={form.submitting}
                    errors={form.submitErrors?.[FORM_ERROR]}
                    clear={form.isFormClearActive}
                    onCancel={onCancel}
                />
            )}
        </LangStatsInsightCreationContent>
    )
}
