import classnames from 'classnames'
import React from 'react'
import { noop } from 'rxjs'

import { FormInput } from '../../../../../components/form/form-input/FormInput'
import { FieldItem } from '../../../../../components/form/hooks/useFieldArray';
import { DataSeries } from '../../../../../core/backend/types'
import { FormColorInput } from '../form-color-input/FormColorInput'

interface FormSeriesInputProps {
    autofocus?: boolean
    /** Enable cancel button. */
    cancel?: boolean
    /** Custom class name for root element of form series. */
    className?: string
    /** On submit handler of series form. */
    onSubmit: () => void
    /** On cancel handler. */
    onCancel?: () => void

    series: FieldItem<DataSeries>
}

/** Displays form series input (three field - name field, query field and color picker). */
export const FormSeriesInput: React.FunctionComponent<FormSeriesInputProps> = props => {
    const { series, className, cancel = false, autofocus = true, onSubmit, onCancel = noop } = props

    const nameField = series.name;
    const queryField = series.query;
    const colorField = series.stroke;

    const hasNameControlledValue = !!nameField.input.value
    const hasQueryControlledValue = !!queryField.input.value

    const handleDone = () => {
        const isSeriesValid = nameField.meta.validState === 'VALID' &&
            queryField.meta.validState === 'VALID' &&
            colorField.meta.validState === 'VALID'

        if (isSeriesValid) {
            onSubmit()
        }
    }

    return (
        <div className={classnames('d-flex flex-column', className)}>
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
                placeholder="Example: patternType:regexp const\s\w+:\s(React\.)?FunctionComponent"
                description={
                    <span>
                        Do not include the <code>repo:</code> filter as it will be added automatically for the
                        repositories you included above.
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
                <button
                    cli
                    aria-label="Submit button for data series"
                    type="button"
                    className="btn btn-secondary"
                >
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
