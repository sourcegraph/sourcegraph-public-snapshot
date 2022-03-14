import classNames from 'classnames'
import React, { useCallback, useEffect } from 'react'

import { asError } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useLocalStorage, Link } from '@sourcegraph/wildcard'

import { PageTitle } from '../../../../../../components/PageTitle'
import { CodeInsightsPage } from '../../../../components/code-insights-page/CodeInsightsPage'
import { FORM_ERROR, FormChangeEvent } from '../../../../components/form/hooks/useForm'
import { LangStatsInsight } from '../../../../core/types'
import { CodeInsightTrackType } from '../../../../pings'

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
     * Whenever the user submit form and clicks on save/submit button
     *
     * @param event - creation event with subject id and updated settings content
     * info.
     */
    onInsightCreateRequest: (event: InsightCreateEvent) => Promise<unknown>

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
    const { telemetryService, onInsightCreateRequest, onCancel, onSuccessfulCreation } = props

    const [initialFormValues, setInitialFormValues] = useLocalStorage<LangStatsCreationFormFields | undefined>(
        'insights.code-stats-creation-ui',
        undefined
    )

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
                    'InsightAddition',
                    { insightType: CodeInsightTrackType.LangStatsInsight },
                    { insightType: CodeInsightTrackType.LangStatsInsight }
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
        <CodeInsightsPage className={classNames(styles.creationPage, 'col-10')}>
            <PageTitle title="Create new code insight" />

            <div className="mb-5">
                <h2>Set up new language usage insight</h2>

                <p className="text-muted">
                    Shows language usage in your repository based on number of lines of code.{' '}
                    <Link to="/help/code_insights" target="_blank" rel="noopener">
                        Learn more.
                    </Link>
                </p>
            </div>

            <LangStatsInsightCreationContent
                className="pb-5"
                initialValues={initialFormValues}
                onSubmit={handleSubmit}
                onCancel={handleCancel}
                onChange={handleChange}
            />
        </CodeInsightsPage>
    )
}
