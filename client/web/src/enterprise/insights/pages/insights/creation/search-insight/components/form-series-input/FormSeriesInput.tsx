import React from 'react'

import classNames from 'classnames'
import { noop } from 'rxjs'

import { Button, Card, Input } from '@sourcegraph/wildcard'

import { getDefaultInputProps } from '../../../../../../components/form/getDefaultInputProps'
import { useField } from '../../../../../../components/form/hooks/useField'
import { useForm } from '../../../../../../components/form/hooks/useForm'
import { InsightQueryInput } from '../../../../../../components/form/query-input/InsightQueryInput'
import { createRequiredValidator } from '../../../../../../components/form/validators'
import { DEFAULT_DATA_SERIES_COLOR } from '../../constants'
import { EditableDataSeries } from '../../types'
import { FormColorInput } from '../form-color-input/FormColorInput'

import { getQueryPatternTypeFilter } from './get-pattern-type-filter'

const requiredNameField = createRequiredValidator('Name is a required field for data series.')
const validQuery = createRequiredValidator('Query is a required field for data series.')

interface FormSeriesInputProps {
    /** Series index. */
    index: number

    /**
     * Show all validation error of all fields within the form.
     */
    showValidationErrorsOnMount?: boolean

    series: EditableDataSeries

    /** Code Insight repositories field string value - repo1, repo2, ... */
    repositories: string

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

export const FormSeriesInput: React.FunctionComponent<React.PropsWithChildren<FormSeriesInputProps>> = props => {
    const {
        index,
        series,
        showValidationErrorsOnMount = false,
        className,
        cancel = false,
        autofocus = true,
        repositories,
        onCancel = noop,
        onSubmit = noop,
        onChange = noop,
    } = props

    const { name, query, stroke: color } = series

    const { formAPI, handleSubmit, ref } = useForm({
        touched: showValidationErrorsOnMount,
        initialValues: {
            seriesName: name ?? '',
            seriesQuery: query ?? '',
            seriesColor: color ?? DEFAULT_DATA_SERIES_COLOR,
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
    })

    const colorField = useField({
        name: 'seriesColor',
        formApi: formAPI,
    })

    return (
        <Card data-testid="series-form" ref={ref} className={classNames('d-flex flex-column', className)}>
            <Input
                label="Name"
                required={true}
                autoFocus={autofocus}
                placeholder="Example: Function component"
                message="Name shown in the legend and tooltip"
                {...getDefaultInputProps(nameField)}
            />

            <Input
                label="Search query"
                required={true}
                as={InsightQueryInput}
                repositories={repositories}
                patternType={getQueryPatternTypeFilter(queryField.input.value)}
                placeholder="Example: patternType:regexp const\s\w+:\s(React\.)?FunctionComponent"
                message={<QueryFieldDescription />}
                className="mt-4"
                {...getDefaultInputProps(queryField)}
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
        </Card>
    )
}

const QueryFieldDescription: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <span>
        Do not include the <code>context:</code> or <code>repo:</code> filter; if needed, <code>repo:</code> will be
        added automatically.
    </span>
)
