import React, { useEffect } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'

import { Page } from '../../../../../../components/Page'
import { PageTitle } from '../../../../../../components/PageTitle'
import { getSanitizedRepositories } from '../../../../components/creation-ui-kit/sanitizers/repositories'
import { FormChangeEvent, SubmissionErrors } from '../../../../components/form/hooks/useForm'
import { CaptureGroupInsight, InsightExecutionType, InsightType } from '../../../../core/types'

import { CaptureGroupCreationContent } from './components/CaptureGroupCreationContent'
import { CaptureGroupFormFields } from './types'

interface CaptureGroupCreationPageProps extends TelemetryProps {
    onInsightCreateRequest: (event: { insight: CaptureGroupInsight }) => Promise<unknown>
    onSuccessfulCreation: (insight: CaptureGroupInsight) => void
    onCancel: () => void
}

export const CaptureGroupCreationPage: React.FunctionComponent<CaptureGroupCreationPageProps> = props => {
    const { telemetryService, onInsightCreateRequest, onSuccessfulCreation, onCancel } = props

    const [initialFormValues, setInitialFormValues] = useLocalStorage<CaptureGroupFormFields | undefined>(
        'insights.code-stats-creation-ui',
        undefined
    )

    useEffect(() => {
        telemetryService.logViewEvent('CodeInsightsCaptureGroupCreationPage')
    }, [telemetryService])

    const handleSubmit = async (values: CaptureGroupFormFields): Promise<SubmissionErrors | void> => {
        const insight = getSanitizedCaptureGroupInsight(values)

        await onInsightCreateRequest({ insight })

        setInitialFormValues(undefined)
        telemetryService.log('CodeInsightsCodeStatsCreationPageSubmitClick')
        telemetryService.log(
            'InsightAddition',
            { insightType: 'codeStatsInsights' },
            { insightType: 'codeStatsInsights' }
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
        <Page>
            <PageTitle title="Create new capture group code insight" />

            <header className="mb-5">
                <h2>Create new code insight</h2>

                <p className="text-muted">
                    Search-based code insights analyze your code based on any search query.{' '}
                    <a href="https://docs.sourcegraph.com/code_insights" target="_blank" rel="noopener">
                        Learn more.
                    </a>
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
        </Page>
    )
}

function getSanitizedCaptureGroupInsight(values: CaptureGroupFormFields): CaptureGroupInsight {
    return {
        title: values.title.trim(),
        query: values.groupSearchQuery.trim(),
        repositories: getSanitizedRepositories(values.repositories),
        viewType: InsightType.CaptureGroup,
        type: InsightExecutionType.Backend,
        id: '',
        visibility: '',
        step: { [values.step]: +values.stepValue },
    }
}
