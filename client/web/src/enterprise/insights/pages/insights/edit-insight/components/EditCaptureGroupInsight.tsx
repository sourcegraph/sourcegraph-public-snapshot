import React, { useMemo } from 'react'

import {
    CodeInsightCreationMode,
    CodeInsightsCreationActions,
    FORM_ERROR,
    SubmissionErrors,
} from '../../../../components'
import { MinimalCaptureGroupInsightData, CaptureGroupInsight } from '../../../../core'
import { CaptureGroupFormFields } from '../../creation/capture-group'
import { CaptureGroupCreationContent } from '../../creation/capture-group/components/CaptureGroupCreationContent'
import { getSanitizedCaptureGroupInsight } from '../../creation/capture-group/utils/capture-group-insight-sanitizer'
import { InsightStep } from '../../creation/search-insight'

interface EditCaptureGroupInsightProps {
    insight: CaptureGroupInsight
    licensed: boolean
    isEditAvailable: boolean | undefined
    onSubmit: (insight: MinimalCaptureGroupInsightData) => SubmissionErrors | Promise<SubmissionErrors> | void
    onCancel: () => void
}

export const EditCaptureGroupInsight: React.FunctionComponent<
    React.PropsWithChildren<EditCaptureGroupInsightProps>
> = props => {
    const { insight, licensed, isEditAvailable, onSubmit, onCancel } = props

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
        return onSubmit({
            ...sanitizedInsight,
            filters: insight.filters,
            seriesDisplayOptions: insight.seriesDisplayOptions,
            seriesCount: insight.seriesCount,
        })
    }

    return (
        <CaptureGroupCreationContent
            touched={true}
            initialValues={insightFormValues}
            className="pb-5"
            onSubmit={handleSubmit}
            onCancel={onCancel}
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
        </CaptureGroupCreationContent>
    )
}
