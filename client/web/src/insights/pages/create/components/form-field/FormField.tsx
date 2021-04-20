import classnames from 'classnames';
import React, { ReactElement } from 'react';

import styles from './FormField.module.scss';

interface InputFieldProps {
    name: string;
    description?: string;
    placeholder?: string;
    className?: string
}

export function InputField(props: InputFieldProps): ReactElement {
    const { name, placeholder, description, className } = props;

    return (
        <label className={classnames(styles.formField, className)}>
            <h4>{name}</h4>

            <input
                type="text"
                className={classnames(styles.formFieldInput, 'form-control')}
                placeholder={placeholder}
            />

            <span className={classnames(styles.formFieldDescription, 'text-muted')}>
                {description}
            </span>
        </label>
    );
}
