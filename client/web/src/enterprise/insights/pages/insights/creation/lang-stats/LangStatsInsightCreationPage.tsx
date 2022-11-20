import { FC, useCallback, useEffect, useMemo } from 'react'

import classNames from 'classnames'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useLocalStorage, Link, PageHeader, useObservable } from '@sourcegraph/wildcard'

import { PageTitle } from '../../../../../../components/PageTitle'
import { CodeInsightsIcon } from '../../../../../../insights/Icons'
import {
    CodeInsightCreationMode,
    CodeInsightsCreationActions,
    CodeInsightsPage,
    FORM_ERROR,
} from '../../../../components'
import { MinimalLangStatsInsightData } from '../../../../core'
import { useUiFeatures } from '../../../../hooks'
import { CodeInsightTrackType } from '../../../../pings'

import {
    LangStatsInsightCreationContent,
    LangStatsInsightCreationContentProps,
} from './components/LangStatsInsightCreationContent'
import { LangStatsCreationFormFields } from './types'
import { getSanitizedLangStatsInsight } from './utils/insight-sanitizer'

import styles from './LangStatsInsightCreationPage.module.scss'

export interface InsightCreateEvent {
    insight: MinimalLangStatsInsightData
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
    onSuccessfulCreation: () => void

    /**
     * Whenever the user click on cancel button
     */
    onCancel: () => void
}

export const LangStatsInsightCreationPage: FC<LangStatsInsightCreationPageProps> = props => {
    const { telemetryService, onInsightCreateRequest, onCancel, onSuccessfulCreation } = props

    const { licensed, insight } = useUiFeatures()
    const creationPermission = useObservable(useMemo(() => insight.getCreationPermissions(), [insight]))

    // We do not use temporal user settings since form values are not so important to
    // waste users time for waiting response of yet another network request to just
    // render creation UI form.
    // eslint-disable-next-line no-restricted-syntax
    const [initialFormValues, setInitialFormValues] = useLocalStorage<LangStatsCreationFormFields | undefined>(
        'insights.code-stats-creation-ui',
        undefined
    )

    useEffect(() => {
        telemetryService.logViewEvent('CodeInsightsCodeStatsCreationPage')
    }, [telemetryService])

    const handleSubmit = useCallback<LangStatsInsightCreationContentProps['onSubmit']>(
        async values => {
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

            onSuccessfulCreation()
        },
        [onInsightCreateRequest, onSuccessfulCreation, setInitialFormValues, telemetryService]
    )

    const handleCancel = useCallback(() => {
        // Clear initial values if user successfully created search insight
        setInitialFormValues(undefined)
        telemetryService.log('CodeInsightsCodeStatsCreationPageCancelClick')

        onCancel()
    }, [setInitialFormValues, telemetryService, onCancel])

    return (
        <CodeInsightsPage className={classNames(styles.creationPage, 'col-10')}>
            <PageTitle title="Create insight - Code Insights" />

            <PageHeader
                className="mb-5"
                path={[{ icon: CodeInsightsIcon }, { text: 'Set up new language usage insight' }]}
                description={
                    <span className="text-muted">
                        Shows language usage in your repository based on number of lines of code.{' '}
                        <Link to="/help/code_insights" target="_blank" rel="noopener">
                            Learn more.
                        </Link>
                    </span>
                }
            />

            <LangStatsInsightCreationContent
                initialValues={initialFormValues}
                touched={false}
                className="pb-5"
                onSubmit={handleSubmit}
                onChange={event => setInitialFormValues(event.values)}
            >
                {form => (
                    <CodeInsightsCreationActions
                        mode={CodeInsightCreationMode.Creation}
                        licensed={licensed}
                        available={creationPermission?.available}
                        submitting={form.submitting}
                        errors={form.submitErrors?.[FORM_ERROR]}
                        clear={form.isFormClearActive}
                        onCancel={handleCancel}
                    />
                )}
            </LangStatsInsightCreationContent>
        </CodeInsightsPage>
    )
}
