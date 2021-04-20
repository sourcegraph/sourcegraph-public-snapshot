import classnames from 'classnames';
import React, { ReactElement } from 'react';

import styles from './FormRadioInput.module.scss'

interface RadioInputProps {
    name: string;
    description?: string;
    className?: string;
}

export function FormRadioInput(props: RadioInputProps): ReactElement {
    const { name, description, className } = props;

    return (
        <label className={classnames(styles.radioInput, className)}>
            <input
                type="radio"
                className={classnames(styles.radioInputInput, 'form-control')}
                required={true}
            />

            <div className={classnames(styles.radioInputDescriptionContent)}>
                <span>{name}</span>
                { description &&
                    <span className='text-muted'>
                        – {description}
                    </span>}
            </div>
        </label>
    );
}
