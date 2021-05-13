import React, { useMemo } from 'react'
import { useHistory } from 'react-router-dom'

import { Settings } from '@sourcegraph/shared/src/settings/settings'

import { SubmissionErrors } from '../../../components/form/hooks/useForm'
import { LangStatsInsight } from '../../../core/types'
import { LangStatsInsightCreationForm } from '../../creation/lang-stats/components/lang-stats-insight-creation-form/LangStatsInsightCreationForm'
import { LangStatsCreationFormFields } from '../../creation/lang-stats/types'
import { getSanitizedLangStatsInsight } from '../../creation/lang-stats/utils/insight-sanitizer'

export interface EditLangStatsInsightProps {
    insight: LangStatsInsight
    onSubmit: (insight: LangStatsInsight) => SubmissionErrors | Promise<SubmissionErrors> | void
    finalSettings: Settings
}

export const EditLangStatsInsight: React.FunctionComponent<EditLangStatsInsightProps> = props => {
    const { insight, finalSettings, onSubmit } = props
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
        <LangStatsInsightCreationForm
            mode="edit"
            className="pb-5"
            initialValues={insightFormValues}
            settings={finalSettings}
            /* eslint-disable-next-line react/jsx-no-bind */
            onSubmit={handleSubmit}
            /* eslint-disable-next-line react/jsx-no-bind */
            onCancel={handleCancel}
        />
    )
}
