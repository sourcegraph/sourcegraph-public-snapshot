import classnames from 'classnames'
import React, { useCallback, useEffect } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { asError } from '@sourcegraph/shared/src/util/errors'

import { Page } from '../../../../../../components/Page'
import { PageTitle } from '../../../../../../components/PageTitle'
import { FORM_ERROR, FormChangeEvent } from '../../../../components/form/hooks/useForm'
import { SearchBasedInsight } from '../../../../core/types'
import { SupportedInsightSubject } from '../../../../core/types/subjects'

import {
    SearchInsightCreationContent,
    SearchInsightCreationContentProps,
} from './components/search-insight-creation-content/SearchInsightCreationContent'
import styles from './SearchInsightCreationPage.module.scss'
import { CreateInsightFormFields } from './types'
import { getSanitizedSearchInsight } from './utils/insight-sanitizer'
import { useSearchInsightInitialValues } from './utils/use-initial-values'

export interface InsightCreateEvent {
    insight: SearchBasedInsight
}

export interface SearchInsightCreationPageProps extends TelemetryProps {
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
    onSuccessfulCreation: (insight: SearchBasedInsight) => void

    /**
     * Whenever the user click on cancel button
     */
    onCancel: () => void
}

/** Displays create insight page with creation form. */
export const SearchInsightCreationPage: React.FunctionComponent<SearchInsightCreationPageProps> = props => {
    const { visibility, subjects, telemetryService, onInsightCreateRequest, onCancel, onSuccessfulCreation } = props

    const { initialValues, loading, setLocalStorageFormValues } = useSearchInsightInitialValues()

    // Set top-level scope value as initial value for the insight visibility
    const mergedInitialValues = { ...initialValues, visibility }

    useEffect(() => {
        telemetryService.logViewEvent('CodeInsightsSearchBasedCreationPage')
    }, [telemetryService])

    const handleSubmit = useCallback<SearchInsightCreationContentProps['onSubmit']>(
        async values => {
            try {
                const insight = getSanitizedSearchInsight(values)

                await onInsightCreateRequest({ insight })

                telemetryService.log('CodeInsightsSearchBasedCreationPageSubmitClick')
                telemetryService.log(
                    'Insight Addition',
                    { insightType: 'searchInsights' },
                    { insightType: 'searchInsights' }
                )

                // Clear initial values if user successfully created search insight
                setLocalStorageFormValues(undefined)

                onSuccessfulCreation(insight)
            } catch (error) {
                return { [FORM_ERROR]: asError(error) }
            }

            return
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
        <Page className={classnames('col-10', styles.creationPage)}>
            <PageTitle title="Create new code insight" />

            {loading && (
                // loading state for 1 click creation insight values resolve operation
                <div>
                    <LoadingSpinner className="icon-inline" /> Resolving search query
                </div>
            )}

            {
                // If we have query in URL we should be sure that we have initial values
                // from URL query based insight. If we don't have query in URl we can render
                // page without resolving URL query based insight values.
                !loading && (
                    <>
                        <div className="mb-5">
                            <h2>Create new code insight</h2>

                            <p className="text-muted">
                                Search-based code insights analyze your code based on any search query.{' '}
                                <a href="https://docs.sourcegraph.com/code_insights" target="_blank" rel="noopener">
                                    Learn more.
                                </a>
                            </p>
                        </div>

                        <SearchInsightCreationContent
                            className="pb-5"
                            dataTestId="search-insight-create-page-content"
                            initialValue={mergedInitialValues}
                            subjects={subjects}
                            onSubmit={handleSubmit}
                            onCancel={handleCancel}
                            onChange={handleChange}
                        />
                    </>
                )
            }
        </Page>
    )
}
