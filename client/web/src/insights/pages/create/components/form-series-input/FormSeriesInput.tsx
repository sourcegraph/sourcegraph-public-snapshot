import classnames from 'classnames';
import React, { ReactElement, Ref, useCallback, useImperativeHandle, useRef } from 'react';
import { useField, useForm } from 'react-final-form-hooks';
import { noop } from 'rxjs';

import { DataSeries } from '../../types';
import { DEFAULT_ACTIVE_COLOR, FormColorInput } from '../form-color-input/FormColorInput';
import { InputField } from '../form-field/FormField';
import { FormGroup } from '../form-group/FormGroup';
import { createRequiredValidator, createValidRegExpValidator, composeValidators } from '../validators';

import styles from './FormSeriesInput.module.scss'

export interface FormSeriesInputAPI {
    focus: () => void
}

interface FormSeriesProps {
    className?: string
    name?: string
    query?: string
    color?: string
    onSubmit?: (series: DataSeries) => void;
    innerRef?: Ref<FormSeriesInputAPI>
}

const requiredNameField = createRequiredValidator('Name is required field for data series.');

const validQuery = composeValidators(
    createValidRegExpValidator('Query must be valid regular expression.'),
    createRequiredValidator('Query is required field for data series.')
)

export function FormSeriesInput(props: FormSeriesProps): ReactElement {
    const { name, query, color, className, onSubmit = noop, innerRef } = props;

    const { handleSubmit, form, } = useForm<DataSeries>({
        initialValues: {
            name,
            query,
            color: color ?? DEFAULT_ACTIVE_COLOR,
        },
        onSubmit
    });

    const nameField = useField('name', form, requiredNameField);
    const queryField = useField('query', form, validQuery)
    const colorField = useField('color', form,)

    const nameReference = useRef<HTMLInputElement>(null);
    const queryReference = useRef<HTMLInputElement>(null);

    // In case if consumer ask this component to be focused (.focus())
    // We focus first invalid field. Otherwise we focus first field of
    // form series - series name field.
    useImperativeHandle(innerRef, () => ({
        focus: () => {
            if (nameField.meta.error) {
                return nameReference.current?.focus()
            }

            if (queryField.meta.error) {
                return queryReference.current?.focus()
            }

            nameReference.current?.focus()
        }
    }));

    const handleSubmitButton = useCallback(
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore
        (event: React.MouseEvent) => handleSubmit(event),
        [handleSubmit]
    )

    return (
        // eslint-disable-next-line react/forbid-elements
        <div className={classnames(styles.formSeriesInput, className)}>

            <InputField
                title='Name'
                placeholder='ex. Function component'
                description='Name shown in the legend and tooltip'
                valid={nameField.meta.touched && nameField.meta.valid}
                error={nameField.meta.touched && nameField.meta.error}
                className={styles.formSeriesInputField}
                {...nameField.input}
                ref={nameReference}
            />

            <InputField
                title='Query'
                placeholder='ex. spatternType:regexp const\\s\\w+:\\s(React\\.)?FunctionComponent'
                description='Do not include the repo: filter as it will be added automatically for the current repository'
                valid={queryField.meta.touched && queryField.meta.valid}
                error={queryField.meta.touched && queryField.meta.error}
                className={styles.formSeriesInputField}
                {...queryField.input}
                ref={queryReference}/>

            <FormGroup
                name='Color'
                className={classnames(styles.formSeriesInputField, styles.formSeriesInputColor)}>

                <FormColorInput
                    value={colorField.input.value}
                    onChange={colorField.input.onChange}
                />
            </FormGroup>

            <div>
                <button
                    type='submit'
                    onClick={handleSubmitButton}
                    className={classnames(styles.formSeriesInputButton,'button')}>

                    Done
                </button>

                <button
                    type='button'
                    className={classnames(styles.formSeriesInputButton, styles.formSeriesInputButtonCancel, 'button')}>

                    Cancel
                </button>
            </div>
        </div>
    );
}
