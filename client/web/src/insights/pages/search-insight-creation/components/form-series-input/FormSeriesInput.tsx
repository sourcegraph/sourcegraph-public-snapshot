import classnames from 'classnames'
import React, { Ref } from 'react'
import { noop } from 'rxjs'

import { useField, useForm } from '../../hooks/useForm'
import { DataSeries } from '../../types'
import { DEFAULT_ACTIVE_COLOR, FormColorInput } from '../form-color-input/FormColorInput'
import { InputField } from '../form-field/FormField'
import { createRequiredValidator, createValidRegExpValidator, composeValidators } from '../validators'

const requiredNameField = createRequiredValidator('Name is required field for data series.')

const validQuery = composeValidators(
    createValidRegExpValidator('Query must be valid regular expression.'),
    createRequiredValidator('Query is required field for data series.')
)

/** Mimic native input public API. */
export interface FormSeriesInputAPI {
    /** Mimic-function to native input focus. */
    focus: () => void
}

interface FormSeriesProps {
    /** Name of series. */
    name?: string
    /** Query value of series. */
    query?: string
    /** Color value for line chart. (series) */
    color?: string
    /** Enable autofocus behavior of first input of form. */
    autofocus?: boolean
    /** Enable cancel button. */
    cancel?: boolean
    /** Ref for mimic native behavior (focus function). */
    innerRef?: Ref<FormSeriesInputAPI>
    /** Custom class name for root element of form series. */
    className?: string
    /** On submit handler of series form. */
    onSubmit?: (series: DataSeries) => void
    /** On cancel handler. */
    onCancel?: () => void
}

/** Displays form series input (three field - name field, query field and color picker). */
export const FormSeriesInput: React.FunctionComponent<FormSeriesProps> = props => {
    const {
        name,
        query,
        color,
        className,
        cancel = false,
        onCancel = noop,
        onSubmit = noop,
    } = props

    const hasNameControlledValue = !!name
    const hasQueryControlledValue = !!query

    const { formAPI, handleSubmit, ref } = useForm({
        initialValues: {
            seriesName: name,
            seriesQuery: query,
            seriesColor: color ?? DEFAULT_ACTIVE_COLOR,
        },
        onSubmit: values =>
            onSubmit({
                name: values.seriesName,
                query: values.seriesQuery,
                color: values.seriesColor,
            }),
    })

    const nameField = useField('seriesName', formAPI, requiredNameField)
    const queryField = useField('seriesQuery', formAPI, validQuery)
    const colorField = useField('seriesColor', formAPI)

    return (
        <div ref={ref} className={classnames('d-flex flex-column', className)}>
            <InputField
                title="Name"
                placeholder="Example: Function component"
                description="Name shown in the legend and tooltip"
                valid={(hasNameControlledValue || nameField.meta.touched) && nameField.meta.validState === 'VALID'}
                error={nameField.meta.touched && nameField.meta.error}
                {...nameField.input}
            />

            <InputField
                title="Query"
                placeholder="Example: patternType:regexp const\s\w+:\s(React\.)?FunctionComponent"
                description={
                    <span>
                        Do not include the <code>repo:</code> filter as it will be added automatically for the current
                        repository
                    </span>
                }
                valid={(hasQueryControlledValue || queryField.meta.touched) && queryField.meta.validState === 'VALID'}
                error={queryField.meta.touched && queryField.meta.error}
                className="mt-4"
                {...queryField.input}
            />

            <FormColorInput
                name="series color group"
                title="Color"
                className="mt-4"
                value={colorField.input.value}
                onChange={colorField.input.onChange}
            />

            <div className="mt-4">
                <button type="button" onClick={handleSubmit} className="btn btn-light">
                    Done
                </button>

                {cancel && (
                    <button type="button" onClick={onCancel} className="btn btn-outline-secondary ml-2">
                        Cancel
                    </button>
                )}
            </div>
        </div>
    )
}
