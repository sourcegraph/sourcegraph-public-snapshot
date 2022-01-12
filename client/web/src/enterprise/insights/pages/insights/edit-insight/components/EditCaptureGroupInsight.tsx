import React, { useMemo } from 'react'

import { SubmissionErrors } from '../../../../components/form/hooks/useForm'
import { CaptureGroupInsight } from '../../../../core/types'
import { CaptureGroupCreationContent } from '../../creation/capture-group/components/CaptureGroupCreationContent'
import { CaptureGroupFormFields } from '../../creation/capture-group/types'
import { getSanitizedCaptureGroupInsight } from '../../creation/capture-group/utils/capture-group-insight-sanitizer'
import { InsightStep } from '../../creation/search-insight/types'

interface EditCaptureGroupInsightProps {
    insight: CaptureGroupInsight
    onSubmit: (insight: CaptureGroupInsight) => SubmissionErrors | Promise<SubmissionErrors> | void
    onCancel: () => void
}

export const EditCaptureGroupInsight: React.FunctionComponent<EditCaptureGroupInsightProps> = props => {
    const { insight, onSubmit, onCancel } = props

    const insightFormValues = useMemo<CaptureGroupFormFields>(
        () => ({
            title: insight.title,
            repositories: insight.repositories.join(', '),
            groupSearchQuery: insight.query,
            stepValue: Object.values(insight.step)[0]?.toString() ?? '3',
            step: Object.keys(insight.step)[0] as InsightStep,
        }),
        [insight]
    )

    const handleSubmit = (values: CaptureGroupFormFields): SubmissionErrors | Promise<SubmissionErrors> | void => {
        const sanitizedInsight = getSanitizedCaptureGroupInsight(values)

        // Preserve backend insight filters since these filters aren't represented
        // in the editing form
        return onSubmit({ ...sanitizedInsight, filters: insight.filters })
    }

    return (
        <CaptureGroupCreationContent
            mode="edit"
            initialValues={insightFormValues}
            onSubmit={handleSubmit}
            onCancel={onCancel}
        />
    )
}
