import classnames from 'classnames';
import React, { ReactElement, useEffect } from 'react';
import { useField, useForm } from 'react-final-form-hooks';

import { InputField } from '../form-field/FormField';
import { FormGroup } from '../form-group/FormGroup';
import { FormColorInput } from '../form-color-input/FormColorInput';

import styles from './FormSeries.module.scss'

interface FormSeriesProps {
    className?: string
    name?: string
    query?: string
    color?: string
}

export function FormSeries(props: FormSeriesProps): ReactElement {
    const { name, query, color, className } = props;

    const form = useForm({
        initialValues: {
            name,
            query,
            color,
        },
        onSubmit: () => console.log('submit')
    });

    const nameField = useField('name', form.form);
    const queryField = useField('query', form.form)
    const colorField = useField('color', form.form)

    // Sync internal value state and form value from top level (props values)
    // it might be useful when we need to pass some value in active series form
    useEffect(() => {
        const {
            name: previousName,
            color: previousColor,
            query: previousQuery,
        } = form.values;

        const values = [
            { id: 'name', nextValue: name, prevValue: previousName },
            { id: 'query', nextValue: query, prevValue: previousQuery },
            { id: 'color', nextValue: color, prevValue: previousColor },
        ] as const;

        for (const field of values) {
            if (field.nextValue !== field.prevValue) {
                form.form.change(field.id, field.nextValue)
            }
        }
    }, [form, name, color, query])

    return (
        <div className={classnames(styles.formSeries, className)}>

            <InputField
                title='Name'
                placeholder='ex. Function component'
                description='Name shown in the legend and tooltip'
                className={styles.formSeriesField}
                {...nameField.input}/>

            <InputField
                title='Query'
                placeholder='ex. spatternType:regexp const\\s\\w+:\\s(React\\.)?FunctionComponent'
                description='Do not include the repo: filter as it will be added automatically for the current repository'
                className={styles.formSeriesField}
                {...queryField.input}/>

            <FormGroup
                name='Color'
                className={styles.formSeriesField}>

                <FormColorInput/>
            </FormGroup>

            <button
                type='button'
                className={classnames(styles.formSeriesButton,'button')}>

                Done
            </button>
        </div>
    );
}
