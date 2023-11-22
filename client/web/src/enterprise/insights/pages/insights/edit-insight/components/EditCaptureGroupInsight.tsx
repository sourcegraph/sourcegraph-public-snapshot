import { useMemo, type FC } from 'react'

import { FORM_ERROR, type SubmissionErrors } from '@sourcegraph/wildcard'

import { CodeInsightCreationMode, CodeInsightsCreationActions } from '../../../../components'
import type { MinimalCaptureGroupInsightData, CaptureGroupInsight } from '../../../../core'
import type { CaptureGroupFormFields } from '../../creation/capture-group'
import { CaptureGroupCreationContent } from '../../creation/capture-group/components/CaptureGroupCreationContent'
import { getSanitizedCaptureGroupInsight } from '../../creation/capture-group/utils/capture-group-insight-sanitizer'
import type { InsightStep } from '../../creation/search-insight'

interface EditCaptureGroupInsightProps {
    insight: CaptureGroupInsight
    licensed: boolean
    isEditAvailable: boolean | undefined
    onSubmit: (insight: MinimalCaptureGroupInsightData) => SubmissionErrors | Promise<SubmissionErrors> | void
    onCancel: () => void
}

export const EditCaptureGroupInsight: FC<EditCaptureGroupInsightProps> = props => {
    const { insight, licensed, isEditAvailable, onSubmit, onCancel } = props

    const insightFormValues = useMemo<CaptureGroupFormFields>(() => {
        const isAllReposInsight = insight.repoQuery === '' && insight.repositories.length === 0
        const repoQuery = isAllReposInsight ? 'repo:.*' : insight.repoQuery

        return {
            title: insight.title,
            repoMode: repoQuery ? 'search-query' : 'urls-list',
            repoQuery: { query: repoQuery },
            repositories: insight.repositories,
            groupSearchQuery: insight.query,
            stepValue: Object.values(insight.step)[0]?.toString() ?? '3',
            step: Object.keys(insight.step)[0] as InsightStep,
            allRepos: insight.repositories.length === 0,
            dashboardReferenceCount: insight.dashboardReferenceCount,
        }
    }, [insight])

    const handleSubmit = (values: CaptureGroupFormFields): SubmissionErrors | Promise<SubmissionErrors> | void => {
        const sanitizedInsight = getSanitizedCaptureGroupInsight(values)

        // Preserve backend insight filters since these filters aren't represented
        // in the editing form
        return onSubmit({
            ...sanitizedInsight,
            filters: insight.filters,
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
