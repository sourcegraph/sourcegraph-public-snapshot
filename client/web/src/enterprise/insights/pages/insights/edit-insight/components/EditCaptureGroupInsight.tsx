import React, { useMemo } from 'react'

import { SubmissionErrors } from '../../../../components/form/hooks/useForm'
import { MinimalCaptureGroupInsightData } from '../../../../core/backend/code-insights-backend-types'
import { CaptureGroupInsight } from '../../../../core/types'
import { CaptureGroupFormFields } from '../../creation/capture-group'
import { CaptureGroupCreationContent } from '../../creation/capture-group/components/CaptureGroupCreationContent'
import { getSanitizedCaptureGroupInsight } from '../../creation/capture-group/utils/capture-group-insight-sanitizer'
import { InsightStep } from '../../creation/search-insight'

interface EditCaptureGroupInsightProps {
    insight: CaptureGroupInsight
    onSubmit: (insight: MinimalCaptureGroupInsightData) => SubmissionErrors | Promise<SubmissionErrors> | void
    onCancel: () => void
}

export const EditCaptureGroupInsight: React.FunctionComponent<
    React.PropsWithChildren<EditCaptureGroupInsightProps>
> = props => {
    const { insight, onSubmit, onCancel } = props

    const insightFormValues = useMemo<CaptureGroupFormFields>(
        () => ({
            title: insight.title,
            repositories: insight.repositories.join(', '),
            groupSearchQuery: insight.query,
            stepValue: Object.values(insight.step)[0]?.toString() ?? '3',
            step: Object.keys(insight.step)[0] as InsightStep,
            allRepos: insight.repositories.length === 0,
            dashboardReferenceCount: insight.dashboardReferenceCount,
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
            className="pb-5"
            insight={insight}
            onSubmit={handleSubmit}
            onCancel={onCancel}
        />
    )
}
