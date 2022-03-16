import React, { useEffect } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Link } from '@sourcegraph/wildcard'

import { PageTitle } from '../../../../../../components/PageTitle'
import { CodeInsightsPage } from '../../../../components/code-insights-page/CodeInsightsPage'
import { FormChangeEvent, SubmissionErrors } from '../../../../components/form/hooks/useForm'
import { CaptureGroupInsight } from '../../../../core/types'
import { CodeInsightTrackType } from '../../../../pings'

import { CaptureGroupCreationContent } from './components/CaptureGroupCreationContent'
import { useCaptureInsightInitialValues } from './hooks/use-capture-insight-initial-values'
import { CaptureGroupFormFields } from './types'
import { getSanitizedCaptureGroupInsight } from './utils/capture-group-insight-sanitizer'

interface CaptureGroupCreationPageProps extends TelemetryProps {
    onInsightCreateRequest: (event: { insight: CaptureGroupInsight }) => Promise<unknown>
    onSuccessfulCreation: (insight: CaptureGroupInsight) => void
    onCancel: () => void
}

export const CaptureGroupCreationPage: React.FunctionComponent<CaptureGroupCreationPageProps> = props => {
    const { telemetryService, onInsightCreateRequest, onSuccessfulCreation, onCancel } = props

    const [initialFormValues, setInitialFormValues] = useCaptureInsightInitialValues()

    useEffect(() => {
        telemetryService.logViewEvent('CodeInsightsCaptureGroupCreationPage')
    }, [telemetryService])

    const handleSubmit = async (values: CaptureGroupFormFields): Promise<SubmissionErrors | void> => {
        const insight = getSanitizedCaptureGroupInsight(values)

        await onInsightCreateRequest({ insight })

        setInitialFormValues(undefined)
        telemetryService.log('CodeInsightsCaptureGroupCreationPageSubmitClick')
        telemetryService.log(
            'InsightAddition',
            { insightType: CodeInsightTrackType.CaptureGroupInsight },
            { insightType: CodeInsightTrackType.CaptureGroupInsight }
        )

        onSuccessfulCreation(insight)
    }

    const handleCancel = (): void => {
        // Clear initial values if user successfully created search insight
        setInitialFormValues(undefined)
        telemetryService.log('CodeInsightsCaptureGroupCreationPageCancelClick')

        onCancel()
    }

    const handleChange = (event: FormChangeEvent<CaptureGroupFormFields>): void => {
        setInitialFormValues(event.values)
    }

    return (
        <CodeInsightsPage>
            <PageTitle title="Create new capture group code insight" />

            <header className="mb-5">
                <h2>Create new code insight</h2>

                <p className="text-muted">
                    Search-based code insights analyze your code based on any search query.{' '}
                    <Link to="/help/code_insights" target="_blank" rel="noopener">
                        Learn more.
                    </Link>
                </p>
            </header>

            <CaptureGroupCreationContent
                mode="creation"
                className="pb-5"
                initialValues={initialFormValues}
                onSubmit={handleSubmit}
                onCancel={handleCancel}
                onChange={handleChange}
            />
        </CodeInsightsPage>
    )
}
