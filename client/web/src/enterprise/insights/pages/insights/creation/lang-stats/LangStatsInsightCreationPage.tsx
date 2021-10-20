import classnames from 'classnames'
import React, { useCallback, useEffect } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { asError } from '@sourcegraph/shared/src/util/errors'
import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'

import { Page } from '../../../../../../components/Page'
import { PageTitle } from '../../../../../../components/PageTitle'
import { FORM_ERROR, FormChangeEvent } from '../../../../components/form/hooks/useForm'
import { LangStatsInsight } from '../../../../core/types'
import { SupportedInsightSubject } from '../../../../core/types/subjects'

import {
    LangStatsInsightCreationContent,
    LangStatsInsightCreationContentProps,
} from './components/lang-stats-insight-creation-content/LangStatsInsightCreationContent'
import styles from './LangStatsInsightCreationPage.module.scss'
import { LangStatsCreationFormFields } from './types'
import { getSanitizedLangStatsInsight } from './utils/insight-sanitizer'

export interface InsightCreateEvent {
    insight: LangStatsInsight
}

export interface LangStatsInsightCreationPageProps extends TelemetryProps {
    /**
     * Set initial value for insight visibility setting.
     */
    visibility: string

    /**
     * List of all supported by code insights subjects that can store insight entities
     * it's used for visibility setting section.
     */
    subjects: SupportedInsightSubject[]

    /**
     * Whenever the user submit form and clicks on save/submit button
     *
     * @param event - creation event with subject id and updated settings content
     * info.
     */
    onInsightCreateRequest: (event: InsightCreateEvent) => Promise<void>

    /**
     * Whenever insight was created and all operations after creation were completed.
     */
    onSuccessfulCreation: (insight: LangStatsInsight) => void

    /**
     * Whenever the user click on cancel button
     */
    onCancel: () => void
}

export const LangStatsInsightCreationPage: React.FunctionComponent<LangStatsInsightCreationPageProps> = props => {
    const { visibility, subjects, telemetryService, onInsightCreateRequest, onCancel, onSuccessfulCreation } = props

    const [initialFormValues, setInitialFormValues] = useLocalStorage<LangStatsCreationFormFields | undefined>(
        'insights.code-stats-creation-ui',
        undefined
    )

    // Set the top-level scope value as initial value for the insight visibility
    const mergedInitialValues = { ...(initialFormValues ?? {}), visibility }

    useEffect(() => {
        telemetryService.logViewEvent('CodeInsightsCodeStatsCreationPage')
    }, [telemetryService])

    const handleSubmit = useCallback<LangStatsInsightCreationContentProps['onSubmit']>(
        async values => {
            try {
                const insight = getSanitizedLangStatsInsight(values)

                await onInsightCreateRequest({ insight })

                // Clear initial values if user successfully created search insight
                setInitialFormValues(undefined)
                telemetryService.log('CodeInsightsCodeStatsCreationPageSubmitClick')
                telemetryService.log(
                    'Insight Addition',
                    { insightType: 'codeStatsInsights' },
                    { insightType: 'codeStatsInsights' }
                )

                onSuccessfulCreation(insight)
            } catch (error) {
                return { [FORM_ERROR]: asError(error) }
            }

            return
        },
        [onInsightCreateRequest, onSuccessfulCreation, setInitialFormValues, telemetryService]
    )

    const handleCancel = useCallback(() => {
        // Clear initial values if user successfully created search insight
        setInitialFormValues(undefined)
        telemetryService.log('CodeInsightsCodeStatsCreationPageCancelClick')

        onCancel()
    }, [setInitialFormValues, telemetryService, onCancel])

    const handleChange = (event: FormChangeEvent<LangStatsCreationFormFields>): void => {
        setInitialFormValues(event.values)
    }

    return (
        <Page className={classnames(styles.creationPage, 'col-10')}>
            <PageTitle title="Create new code insight" />

            <div className="mb-5">
                <h2>Set up new language usage insight</h2>

                <p className="text-muted">
                    Shows language usage in your repository based on number of lines of code.{' '}
                    <a href="https://docs.sourcegraph.com/code_insights" target="_blank" rel="noopener">
                        Learn more.
                    </a>
                </p>
            </div>

            <LangStatsInsightCreationContent
                className="pb-5"
                initialValues={mergedInitialValues}
                subjects={subjects}
                onSubmit={handleSubmit}
                onCancel={handleCancel}
                onChange={handleChange}
            />
        </Page>
    )
}
