import classnames from 'classnames'
import React, { Ref, useCallback, useEffect, useImperativeHandle, useRef } from 'react'
import { useField, useForm } from 'react-final-form-hooks'
import { noop } from 'rxjs'

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

interface FormSeriesValues {
    seriesName: string
    seriesQuery: string
    seriesColor: string
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
        autofocus = false,
        onCancel = noop,
        onSubmit = noop,
        innerRef,
    } = props

    const hasNameControlledValue = !!name
    const hasQueryControlledValue = !!query

    const { handleSubmit, form } = useForm<FormSeriesValues>({
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

    const nameField = useField('seriesName', form, requiredNameField)
    const queryField = useField('seriesQuery', form, validQuery)
    const colorField = useField('seriesColor', form)

    const nameReference = useRef<HTMLInputElement>(null)
    const queryReference = useRef<HTMLInputElement>(null)

    // In case if consumer asked this component to be focused (call .focus() on ref)
    // We focus first invalid field. Otherwise we focus first field of form series
    // - series name field.
    useImperativeHandle(innerRef, () => ({
        focus: () => {
            if (nameField.meta.error) {
                return nameReference.current?.focus()
            }

            if (queryField.meta.error) {
                return queryReference.current?.focus()
            }

            nameReference.current?.focus()
        },
    }))

    useEffect(() => {
        if (autofocus) {
            nameReference.current?.focus()
        }
    }, [autofocus])

    const handleSubmitButton = useCallback(
        async (event: React.MouseEvent) => {

            // handleSubmit work with form element and use form event
            // but we can't have sub forms for the sake of semantics.
            // if this case synthetic event of totally comparable with event
            // from submit button.
            await handleSubmit((event as unknown) as React.SyntheticEvent<HTMLFormElement>)

            if (nameField.meta.error) {
                event.preventDefault()
                return nameReference.current?.focus()
            }

            if (queryField.meta.error) {
                event.preventDefault()
                return queryReference.current?.focus()
            }
        },
        [handleSubmit, nameField.meta.error, queryField.meta.error]
    )

    return (
        <div className={classnames('d-flex flex-column', className)}>
            <InputField
                title="Name"
                placeholder="ex. Function component"
                description="Name shown in the legend and tooltip"
                valid={(hasNameControlledValue || nameField.meta.touched) && nameField.meta.valid}
                error={nameField.meta.touched && nameField.meta.error}
                {...nameField.input}
                ref={nameReference}
            />

            <InputField
                title="Query"
                placeholder="ex. spatternType:regexp const\\s\\w+:\\s(React\\.)?FunctionComponent"
                description="Do not include the repo: filter as it will be added automatically for the current repository"
                valid={(hasQueryControlledValue || queryField.meta.touched) && queryField.meta.valid}
                error={queryField.meta.touched && queryField.meta.error}
                className='mt-4'
                {...queryField.input}
                ref={queryReference}
            />

            <FormColorInput
                name="series color group"
                title="Color"
                className='mt-4'
                value={colorField.input.value}
                onChange={colorField.input.onChange}
            />

            <div className='mt-4'>
                <button
                    type="submit"
                    onClick={handleSubmitButton}
                    className='btn btn-light'
                >
                    Done
                </button>

                {cancel && (
                    <button
                        type="button"
                        onClick={onCancel}
                        className='btn btn-outline-secondary ml-2'
                    >
                        Cancel
                    </button>
                )}
            </div>
        </div>
    )
}
