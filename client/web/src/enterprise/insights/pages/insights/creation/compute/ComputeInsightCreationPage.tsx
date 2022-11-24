import { FunctionComponent, useCallback, useMemo } from 'react'

import BarChartIcon from 'mdi-react/BarChartIcon'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Link, PageHeader, Text, useLocalStorage, useObservable } from '@sourcegraph/wildcard'

import { PageTitle } from '../../../../../../components/PageTitle'
import {
    CodeInsightCreationMode,
    CodeInsightsCreationActions,
    CodeInsightsPage,
    FORM_ERROR,
    FormChangeEvent,
    SubmissionErrors,
} from '../../../../components'
import { ComputeInsight } from '../../../../core'
import { useUiFeatures } from '../../../../hooks'
import { CodeInsightTrackType } from '../../../../pings'

import { ComputeInsightCreationContent } from './components/ComputeInsightCreationContent'
import { CreateComputeInsightFormFields } from './types'
import { getSanitizedComputeInsight } from './utils/insight-sanitaizer'

export interface InsightCreateEvent {
    insight: ComputeInsight
}

interface ComputeInsightCreationPageProps extends TelemetryProps {
    onInsightCreateRequest: (event: InsightCreateEvent) => Promise<unknown>
    onSuccessfulCreation: () => void
    onCancel: () => void
}

export const ComputeInsightCreationPage: FunctionComponent<ComputeInsightCreationPageProps> = props => {
    const { telemetryService, onInsightCreateRequest, onSuccessfulCreation, onCancel } = props

    const { licensed, insight } = useUiFeatures()
    const creationPermission = useObservable(useMemo(() => insight.getCreationPermissions(), [insight]))

    // We do not use temporal user settings since form values are not so important to
    // waste users time for waiting response of yet another network request to just
    // render creation UI form.
    // eslint-disable-next-line no-restricted-syntax
    const [initialFormValues, setInitialFormValues] = useLocalStorage<CreateComputeInsightFormFields | undefined>(
        'insights.compute-creation-ui',
        undefined
    )

    const handleChange = (event: FormChangeEvent<CreateComputeInsightFormFields>): void => {
        setInitialFormValues(event.values)
    }

    const handleSubmit = useCallback(
        async (values: CreateComputeInsightFormFields): Promise<SubmissionErrors> => {
            const insight = getSanitizedComputeInsight(values)

            await onInsightCreateRequest({ insight })

            // Clear initial values if user successfully created search insight
            setInitialFormValues(undefined)
            telemetryService.log('CodeInsightsComputeCreationPageSubmitClick')
            telemetryService.log(
                'InsightAddition',
                { insightType: CodeInsightTrackType.ComputeInsight },
                { insightType: CodeInsightTrackType.ComputeInsight }
            )

            onSuccessfulCreation()
        },
        [onInsightCreateRequest, onSuccessfulCreation, setInitialFormValues, telemetryService]
    )

    const handleCancel = useCallback(() => {
        // Clear initial values if user successfully created search insight
        setInitialFormValues(undefined)
        telemetryService.log('CodeInsightsComputeCreationPageCancelClick')

        onCancel()
    }, [setInitialFormValues, telemetryService, onCancel])

    return (
        <CodeInsightsPage className="col-11">
            <PageTitle title="Create compute insight - Code Insights" />

            <PageHeader
                className="mb-5"
                path={[{ icon: BarChartIcon }, { text: 'Create code insight' }]}
                description={
                    <Text>
                        Type: <b>Group results</b> | <Link to="/insights/create">Change type</Link>
                    </Text>
                }
            />

            <ComputeInsightCreationContent
                touched={false}
                initialValue={initialFormValues}
                data-testid="search-insight-create-page-content"
                className="pb-5"
                onChange={handleChange}
                onSubmit={handleSubmit}
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
            </ComputeInsightCreationContent>
        </CodeInsightsPage>
    )
}
