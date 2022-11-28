import { FC, useCallback, useEffect, useMemo } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner, Link, PageHeader, useObservable } from '@sourcegraph/wildcard'

import { PageTitle } from '../../../../../../components/PageTitle'
import { CodeInsightsIcon } from '../../../../../../insights/Icons'
import {
    FORM_ERROR,
    FormChangeEvent,
    CodeInsightsPage,
    CodeInsightsCreationActions,
    CodeInsightCreationMode,
} from '../../../../components'
import { MinimalSearchBasedInsightData } from '../../../../core'
import { useUiFeatures } from '../../../../hooks'
import { CodeInsightTrackType } from '../../../../pings'

import {
    SearchInsightCreationContent,
    SearchInsightCreationContentProps,
} from './components/SearchInsightCreationContent'
import { CreateInsightFormFields } from './types'
import { getSanitizedSearchInsight } from './utils/insight-sanitizer'
import { useSearchInsightInitialValues } from './utils/use-initial-values'

import styles from './SearchInsightCreationPage.module.scss'

export interface InsightCreateEvent {
    insight: MinimalSearchBasedInsightData
}

export interface SearchInsightCreationPageProps extends TelemetryProps {
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

export const SearchInsightCreationPage: FC<SearchInsightCreationPageProps> = props => {
    const { telemetryService, onInsightCreateRequest, onCancel, onSuccessfulCreation } = props

    const { licensed, insight } = useUiFeatures()
    const creationPermission = useObservable(useMemo(() => insight.getCreationPermissions(), [insight]))

    const { initialValues, loading, setLocalStorageFormValues } = useSearchInsightInitialValues()

    useEffect(() => {
        telemetryService.logViewEvent('CodeInsightsSearchBasedCreationPage')
    }, [telemetryService])

    const handleSubmit = useCallback<SearchInsightCreationContentProps['onSubmit']>(
        async values => {
            const insight = getSanitizedSearchInsight(values)

            await onInsightCreateRequest({ insight })

            telemetryService.log('CodeInsightsSearchBasedCreationPageSubmitClick')
            telemetryService.log(
                'InsightAddition',
                { insightType: CodeInsightTrackType.SearchBasedInsight },
                { insightType: CodeInsightTrackType.SearchBasedInsight }
            )

            // Clear initial values if user successfully created search insight
            setLocalStorageFormValues(undefined)

            onSuccessfulCreation()
        },
        [onInsightCreateRequest, telemetryService, setLocalStorageFormValues, onSuccessfulCreation]
    )

    const handleChange = (event: FormChangeEvent<CreateInsightFormFields>): void => {
        setLocalStorageFormValues(event.values)
    }

    const handleCancel = useCallback(() => {
        telemetryService.log('CodeInsightsSearchBasedCreationPageCancelClick')
        setLocalStorageFormValues(undefined)
        onCancel()
    }, [telemetryService, setLocalStorageFormValues, onCancel])

    return (
        <CodeInsightsPage className={styles.creationPage}>
            <PageTitle title="Create insight - Code Insights" />

            {loading && (
                // loading state for 1 click creation insight values resolve operation
                <div>
                    <LoadingSpinner /> Resolving search query
                </div>
            )}

            {
                // If we have a query in URL we should be sure that we have initial values
                // from URL query based insight. If we don't have query in URl we can render
                // page without resolving URL query based insight values.
                !loading && (
                    <>
                        <PageHeader
                            className="mb-5"
                            path={[{ icon: CodeInsightsIcon }, { text: 'Create new code insight' }]}
                            description={
                                <span className="text-muted">
                                    Search-based code insights analyze your code based on any search query.{' '}
                                    <Link to="/help/code_insights" target="_blank" rel="noopener">
                                        Learn more.
                                    </Link>
                                </span>
                            }
                        />

                        <SearchInsightCreationContent
                            touched={false}
                            initialValue={initialValues}
                            dataTestId="search-insight-create-page-content"
                            className="pb-5"
                            onSubmit={handleSubmit}
                            onChange={handleChange}
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
                        </SearchInsightCreationContent>
                    </>
                )
            }
        </CodeInsightsPage>
    )
}
