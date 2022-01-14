import classNames from 'classnames'
import React from 'react'
import { noop } from 'rxjs'

import { Button } from '@sourcegraph/wildcard'

import { SearchPatternType } from '../../../../../../../../graphql-operations'
import { FormInput } from '../../../../../../components/form/form-input/FormInput'
import { useField } from '../../../../../../components/form/hooks/useField'
import { useForm } from '../../../../../../components/form/hooks/useForm'
import { InsightQueryInput } from '../../../../../../components/form/query-input/InsightQueryInput'
import { createRequiredValidator } from '../../../../../../components/form/validators'
import { EditableDataSeries } from '../../types'
import { DEFAULT_ACTIVE_COLOR, FormColorInput } from '../form-color-input/FormColorInput'

const requiredNameField = createRequiredValidator('Name is a required field for data series.')
const validQuery = createRequiredValidator('Query is a required field for data series.')

interface FormSeriesInputProps {
    /** Series index. */
    index: number

    /**
     * This prop represents the case whenever the edit insight UI page
     * deals with backend insight. We need to disable our search insight
     * query field since our backend insight can't update BE data according
     * to the latest insight configuration.
     */
    isSearchQueryDisabled: boolean

    /**
     * Show all validation error of all fields within the form.
     */
    showValidationErrorsOnMount?: boolean

    series: EditableDataSeries

    /** Enable autofocus behavior of first input of form. */
    autofocus?: boolean

    /** Enable cancel button. */
    cancel?: boolean

    /** Custom class name for root element of form series. */
    className?: string

    /** Whenever a user clicks submit (done) button of the series form. */
    onSubmit?: (series: EditableDataSeries) => void

    /** Whenever a user clicks cancel button of the series form. */
    onCancel?: () => void

    /** Whenever a user types new values in any field of the series form. */
    onChange?: (formValues: EditableDataSeries, valid: boolean) => void
}

export const FormSeriesInput: React.FunctionComponent<FormSeriesInputProps> = props => {
    const {
        index,
        series,
        isSearchQueryDisabled,
        showValidationErrorsOnMount = false,
        className,
        cancel = false,
        autofocus = true,
        onCancel = noop,
        onSubmit = noop,
        onChange = noop,
    } = props

    const { name, query, stroke: color } = series
    const hasNameControlledValue = !!name
    const hasQueryControlledValue = !!query

    const { formAPI, handleSubmit, ref } = useForm({
        touched: showValidationErrorsOnMount,
        initialValues: {
            seriesName: name ?? '',
            seriesQuery: query ?? '',
            seriesColor: color ?? DEFAULT_ACTIVE_COLOR,
        },
        onSubmit: values =>
            onSubmit({
                ...series,
                name: values.seriesName,
                query: values.seriesQuery,
                stroke: values.seriesColor,
            }),
        onChange: event => {
            const { values } = event

            onChange(
                {
                    ...series,
                    name: values.seriesName,
                    query: values.seriesQuery,
                    stroke: values.seriesColor,
                },
                event.valid
            )
        },
    })

    const nameField = useField({
        name: 'seriesName',
        formApi: formAPI,
        validators: { sync: requiredNameField },
    })

    const queryField = useField({
        name: 'seriesQuery',
        formApi: formAPI,
        validators: { sync: validQuery },
        disabled: isSearchQueryDisabled,
    })

    const colorField = useField({
        name: 'seriesColor',
        formApi: formAPI,
    })

    return (
        <div data-testid="series-form" ref={ref} className={classNames('d-flex flex-column', className)}>
            <FormInput
                title="Name"
                required={true}
                autoFocus={autofocus}
                placeholder="Example: Function component"
                description="Name shown in the legend and tooltip"
                valid={(hasNameControlledValue || nameField.meta.touched) && nameField.meta.validState === 'VALID'}
                error={nameField.meta.touched && nameField.meta.error}
                {...nameField.input}
            />

            <FormInput
                title="Search query"
                required={true}
                as={InsightQueryInput}
                patternType={SearchPatternType.literal}
                placeholder="Example: patternType:regexp const\s\w+:\s(React\.)?FunctionComponent"
                description={<QueryFieldDescription isSearchQueryDisabled={isSearchQueryDisabled} />}
                valid={(hasQueryControlledValue || queryField.meta.touched) && queryField.meta.validState === 'VALID'}
                error={queryField.meta.touched && queryField.meta.error}
                className="mt-4"
                {...queryField.input}
            />

            <FormColorInput
                name={`color group of ${index} series`}
                title="Color"
                className="mt-4"
                value={colorField.input.value}
                onChange={colorField.input.onChange}
            />

            <div className="mt-4">
                <Button
                    aria-label="Submit button for data series"
                    type="button"
                    variant="secondary"
                    onClick={handleSubmit}
                >
                    Done
                </Button>

                {cancel && (
                    <Button type="button" onClick={onCancel} variant="secondary" outline={true} className="ml-2">
                        Cancel
                    </Button>
                )}
            </div>
        </div>
    )
}

const QueryFieldDescription: React.FunctionComponent<{ isSearchQueryDisabled: boolean }> = props => (
    <span>
        {!props.isSearchQueryDisabled ? (
            <>
                Do not include the <code>context:</code> or <code>repo:</code> filter; if needed, <code>repo:</code>{' '}
                will be added automatically.
                <br />
                Tip: include <code>archived:no</code> and <code>fork:no</code> if you don't want results from archived
                or forked repos.
            </>
        ) : (
            <>
                We don't yet allow editing queries for insights over all repos. To change the query, make a new insight.
                This is a known{' '}
                <a
                    href="https://docs.sourcegraph.com/code_insights/explanations/current_limitations_of_code_insights"
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    beta limitation
                </a>
            </>
        )}
    </span>
)
