import { type ChangeEvent, type FocusEventHandler, forwardRef } from 'react'

import classNames from 'classnames'

import { getInputStatus, Input, FormGroup } from '@sourcegraph/wildcard'

import type { InsightStep } from '../../../pages/insights/creation/search-insight'
import { FormRadioInput } from '../../form/form-radio-input/FormRadioInput'

import { getDescriptionText } from './get-interval-descrtiption-text/get-interval-description-text'

import styles from './CodeInsightTimeStepPicker.module.scss'

interface CodeInsightTimeStepPickerProps {
    value: string | number
    numberOfPoints: number
    name?: string
    valid?: boolean
    disabled?: boolean
    error?: string
    errorInputState?: boolean

    onChange: (event: ChangeEvent<HTMLInputElement>) => void
    onFocus?: FocusEventHandler<HTMLInputElement>
    onBlur: FocusEventHandler<HTMLInputElement>

    stepType: InsightStep
    onStepTypeChange: (event: ChangeEvent<HTMLInputElement>) => void
}

export const CodeInsightTimeStepPicker = forwardRef<HTMLInputElement, CodeInsightTimeStepPickerProps>(
    (props, reference) => {
        const {
            error,
            errorInputState,
            valid,
            disabled,
            name,
            value,
            stepType,
            numberOfPoints,
            onChange,
            onStepTypeChange,
            onBlur,
            onFocus,
        } = props

        return (
            <FormGroup
                name="insight step group"
                title="Granularity: distance between data points"
                description={getDescriptionText({ stepValue: +value, stepType, numberOfPoints })}
                error={error}
                className="mt-4"
                labelClassName={styles.groupLabel}
                contentClassName="d-flex flex-wrap mb-n2"
            >
                <Input
                    placeholder="ex. 2"
                    required={true}
                    type="number"
                    min={1}
                    disabled={disabled}
                    ref={reference}
                    name={name}
                    value={value}
                    onChange={onChange}
                    onBlur={onBlur}
                    onFocus={onFocus}
                    className={classNames(styles.stepInput)}
                    status={getInputStatus({
                        isValid: valid,
                        isError: errorInputState,
                    })}
                />

                <FormRadioInput
                    title="Hours"
                    name="step"
                    value="hours"
                    checked={stepType === 'hours'}
                    onChange={onStepTypeChange}
                    disabled={disabled}
                    className="mr-3"
                />
                <FormRadioInput
                    title="Days"
                    name="step"
                    value="days"
                    checked={stepType === 'days'}
                    onChange={onStepTypeChange}
                    disabled={disabled}
                    className="mr-3"
                />
                <FormRadioInput
                    title="Weeks"
                    name="step"
                    value="weeks"
                    checked={stepType === 'weeks'}
                    onChange={onStepTypeChange}
                    disabled={disabled}
                    className="mr-3"
                />
                <FormRadioInput
                    title="Months"
                    name="step"
                    value="months"
                    checked={stepType === 'months'}
                    onChange={onStepTypeChange}
                    disabled={disabled}
                    className="mr-3"
                />
                <FormRadioInput
                    title="Years"
                    name="step"
                    value="years"
                    checked={stepType === 'years'}
                    onChange={onStepTypeChange}
                    disabled={disabled}
                    className="mr-3"
                />
            </FormGroup>
        )
    }
)
