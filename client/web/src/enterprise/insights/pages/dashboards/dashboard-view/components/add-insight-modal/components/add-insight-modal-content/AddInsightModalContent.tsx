import React from 'react'

import { escapeRegExp } from 'lodash'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Button, Input, Link, Label, Checkbox } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../../../../../../components/LoaderButton'
import { AccessibleInsight } from '../../../../../../../../../graphql-operations'
import {
    TruncatedText,
    useCheckboxes,
    useField,
    SubmissionErrors,
    useForm,
    FORM_ERROR,
} from '../../../../../../../components'
import { encodeDashboardIdQueryParam } from '../../../../../../../routers.constant'

import styles from './AddInsightModalContent.module.scss'

interface AddInsightModalContentProps {
    insights: AccessibleInsight[]
    initialValues: AddInsightFormValues
    dashboardID: string
    onSubmit: (values: AddInsightFormValues) => SubmissionErrors | Promise<SubmissionErrors> | void
    onCancel: () => void
}

export interface AddInsightFormValues {
    searchInput: string
    insightIds: string[]
}

export const AddInsightModalContent: React.FunctionComponent<
    React.PropsWithChildren<AddInsightModalContentProps>
> = props => {
    const { initialValues, insights, dashboardID, onSubmit, onCancel } = props

    const { formAPI, ref, handleSubmit } = useForm({
        initialValues,
        onSubmit,
    })

    const searchInput = useField({
        name: 'searchInput',
        formApi: formAPI,
    })

    const {
        input: { isChecked, onChange, onBlur },
    } = useCheckboxes('insightIds', formAPI)

    const filteredInsights = insights.filter(insight =>
        insight.presentation.title.match(new RegExp(escapeRegExp(searchInput.input.value), 'gi'))
    )

    return (
        // eslint-disable-next-line react/forbid-elements
        <form ref={ref} onSubmit={handleSubmit}>
            <Input
                autoFocus={true}
                message={
                    <span className="">
                        Don't see an insight? Check the insight's visibility settings or{' '}
                        <Link to={encodeDashboardIdQueryParam('/insights/create', dashboardID)}>
                            create a new insight
                        </Link>
                    </span>
                }
                placeholder="Search insights..."
                {...searchInput.input}
            />

            <fieldset className={styles.insightsContainer}>
                {filteredInsights.map(insight => (
                    <Label key={insight.id} weight="medium" className={styles.insightItem}>
                        <Checkbox
                            name="insightIds"
                            value={insight.id}
                            checked={isChecked(insight.id)}
                            onChange={onChange}
                            onBlur={onBlur}
                            aria-labelledby={insight.id}
                            className={styles.checkbox}
                            wrapperClassName={styles.checkboxWrapper}
                        />
                        <TruncatedText id={insight.id}>{insight.presentation.title}</TruncatedText>
                    </Label>
                ))}
            </fieldset>

            <hr />

            {formAPI.submitErrors?.[FORM_ERROR] && (
                <ErrorAlert className="mt-3" error={formAPI.submitErrors[FORM_ERROR]} />
            )}

            <div className="d-flex justify-content-end mt-4">
                <Button type="button" className="mr-2" variant="secondary" outline={true} onClick={onCancel}>
                    Cancel
                </Button>

                <LoaderButton
                    alwaysShowLabel={true}
                    loading={formAPI.submitting}
                    label={formAPI.submitting ? 'Saving' : 'Save'}
                    type="submit"
                    disabled={formAPI.submitting}
                    variant="primary"
                />
            </div>
        </form>
    )
}
