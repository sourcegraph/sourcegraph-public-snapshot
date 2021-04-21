import classnames from 'classnames';
import React, { ReactElement } from 'react';
import { useField, useForm } from 'react-final-form-hooks';

import { DEFAULT_ACTIVE_COLOR, FormColorInput } from '../form-color-input/FormColorInput';
import { InputField } from '../form-field/FormField';
import { FormGroup } from '../form-group/FormGroup';
import { createRequiredValidator, createValidRegExpValidator, composeValidators } from '../validators';

import styles from './FormSeriesInput.module.scss'

interface FormSeriesProps {
    className?: string
    name?: string
    query?: string
    color?: string
}

const requiredNameField = createRequiredValidator('Name is required field for data series.');

const validQuery = composeValidators(
    createValidRegExpValidator('Query must be valid regular expression.'),
    createRequiredValidator('Query is required field for data series.')
)

export function FormSeriesInput(props: FormSeriesProps): ReactElement {
    const { name, query, color, className } = props;

    const form = useForm({
        initialValues: {
            name,
            query,
            color: color ?? DEFAULT_ACTIVE_COLOR,
        },
        onSubmit: () => console.log('submit')
    });

    const nameField = useField('name', form.form, requiredNameField);
    const queryField = useField('query', form.form, validQuery)
    const colorField = useField('color', form.form,)

    return (
        // eslint-disable-next-line react/forbid-elements
        <form onSubmit={form.handleSubmit} className={classnames(styles.formSeriesInput, className)}>

            <InputField
                title='Name'
                placeholder='ex. Function component'
                description='Name shown in the legend and tooltip'
                error={nameField.meta.touched && nameField.meta.error}
                className={styles.formSeriesInputField}
                {...nameField.input}/>

            <InputField
                title='Query'
                placeholder='ex. spatternType:regexp const\\s\\w+:\\s(React\\.)?FunctionComponent'
                description='Do not include the repo: filter as it will be added automatically for the current repository'
                error={queryField.meta.touched && queryField.meta.error}
                className={styles.formSeriesInputField}
                {...queryField.input}/>

            <FormGroup
                name='Color'
                className={styles.formSeriesInputField}>

                <FormColorInput
                    value={colorField.input.value}
                    onChange={colorField.input.onChange}
                />
            </FormGroup>

            <button
                type='submit'
                className={classnames(styles.formSeriesInputButton,'button')}>

                Done
            </button>

            <pre>{JSON.stringify(form.values,)}</pre>
        </form>
    );
}
