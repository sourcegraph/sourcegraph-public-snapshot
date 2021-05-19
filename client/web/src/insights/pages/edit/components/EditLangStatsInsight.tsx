import React, { useMemo } from 'react'
import { useHistory } from 'react-router-dom'

import { Settings } from '@sourcegraph/shared/src/settings/settings'

import { SubmissionErrors } from '../../../components/form/hooks/useForm'
import { Organization } from '../../../components/visibility-picker/VisibilityPicker'
import { LangStatsInsight } from '../../../core/types'
import { LangStatsInsightCreationContent } from '../../creation/lang-stats/components/lang-stats-insight-creation-content/LangStatsInsightCreationContent'
import { LangStatsCreationFormFields } from '../../creation/lang-stats/types'
import { getSanitizedLangStatsInsight } from '../../creation/lang-stats/utils/insight-sanitizer'

export interface EditLangStatsInsightProps {
    insight: LangStatsInsight
    onSubmit: (insight: LangStatsInsight) => SubmissionErrors | Promise<SubmissionErrors> | void
    finalSettings: Settings
    organizations: Organization[]
}

export const EditLangStatsInsight: React.FunctionComponent<EditLangStatsInsightProps> = props => {
    const { insight, finalSettings, organizations, onSubmit } = props
    const history = useHistory()

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

    const handleCancel = (): void => {
        history.push('/insights')
    }

    return (
        <LangStatsInsightCreationContent
            mode="edit"
            className="pb-5"
            initialValues={insightFormValues}
            settings={finalSettings}
            organizations={organizations}
            onSubmit={handleSubmit}
            onCancel={handleCancel}
        />
    )
}
