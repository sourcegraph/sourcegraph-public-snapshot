import React, { useMemo } from 'react'
import { useHistory } from 'react-router-dom'

import { Settings } from '@sourcegraph/shared/src/settings/settings'

import { SubmissionErrors } from '../../../components/form/hooks/useForm'
import { Organization } from '../../../components/visibility-picker/VisibilityPicker'
import { SearchBasedInsight } from '../../../core/types'
import { SearchInsightCreationContent } from '../../creation/search-insight/components/search-insight-creation-content/SearchInsightCreationContent'
import { CreateInsightFormFields, InsightStep } from '../../creation/search-insight/types'
import { getSanitizedSearchInsight } from '../../creation/search-insight/utils/insight-sanitizer'

interface EditSearchBasedInsightProps {
    insight: SearchBasedInsight
    onSubmit: (insight: SearchBasedInsight) => SubmissionErrors | Promise<SubmissionErrors> | void
    finalSettings: Settings
    organizations: Organization[]
}

export const EditSearchBasedInsight: React.FunctionComponent<EditSearchBasedInsightProps> = props => {
    const { insight, finalSettings = {}, organizations, onSubmit } = props
    const history = useHistory()

    const insightFormValues = useMemo<CreateInsightFormFields>(
        () => ({
            title: insight.title,
            visibility: insight.visibility,
            repositories: insight.repositories.join(', '),
            series: insight.series,
            stepValue: Object.values(insight.step)[0]?.toString() ?? '3',
            step: Object.keys(insight.step)[0] as InsightStep,
        }),
        [insight]
    )

    // Handlers
    const handleSubmit = (values: CreateInsightFormFields): SubmissionErrors | Promise<SubmissionErrors> | void => {
        const sanitizedInsight = getSanitizedSearchInsight(values)

        return onSubmit(sanitizedInsight)
    }

    const handleCancel = (): void => {
        history.push('/insights')
    }

    return (
        <SearchInsightCreationContent
            mode="edit"
            className="pb-5"
            initialValue={insightFormValues}
            settings={finalSettings}
            organizations={organizations}
            onSubmit={handleSubmit}
            onCancel={handleCancel}
        />
    )
}
