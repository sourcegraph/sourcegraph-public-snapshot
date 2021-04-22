import classnames from 'classnames';
import React, { InputHTMLAttributes, ReactElement } from 'react';

import styles from './FormRadioInput.module.scss'

interface RadioInputProps extends InputHTMLAttributes<HTMLInputElement> {
    title: string;
    description?: string;
    className?: string;
}

export function FormRadioInput(props: RadioInputProps): ReactElement {
    const { title, description, className, ...otherProps } = props;

    return (
        <label className={classnames(styles.radioInput, className)}>
            <input
                type="radio"
                className={classnames(styles.radioInputInput, 'form-control')}
                {...otherProps}
            />

            <div className={classnames(styles.radioInputDescriptionContent)}>
                <span>{title}</span>
                { description &&
                    <span className='text-muted'>
                        {' '} – {description}
                    </span>}
            </div>
        </label>
    );
}
