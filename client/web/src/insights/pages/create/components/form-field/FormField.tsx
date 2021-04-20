import classnames from 'classnames';
import React, { InputHTMLAttributes, ReactElement } from 'react';

import styles from './FormField.module.scss';

interface InputFieldProps extends InputHTMLAttributes<HTMLInputElement>{
    title: string;
    description?: string;
    className?: string
}

export function InputField(props: InputFieldProps): ReactElement {
    const { title, placeholder, description, className, ...otherProps } = props;

    return (
        <label className={classnames(styles.formField, className)}>
            <h4>{title}</h4>

            <input
                type="text"
                className={classnames(styles.formFieldInput, 'form-control')}
                {...otherProps}
            />

            <span className={classnames(styles.formFieldDescription, 'text-muted')}>
                {description}
            </span>
        </label>
    );
}
