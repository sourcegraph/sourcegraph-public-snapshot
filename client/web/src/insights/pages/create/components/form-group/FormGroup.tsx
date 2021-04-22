import classnames from 'classnames';
import React, { PropsWithChildren, ReactElement } from 'react';

import styles from './FormGroup.module.scss'

interface FormGroupProps {
    className?: string;
    name: string;
    subtitle?: string;
    error?: string;
    description?: string;
}

export function FormGroup(props: PropsWithChildren<FormGroupProps>): ReactElement {
    const { className, name, subtitle, children, description, error } = props;

    return (
        <fieldset className={classnames(styles.formGroup, className, {[styles.formGroupWithSubtitle]: !!subtitle })}>

            <div className={styles.formGroupNameBlock}>
                <h4 className={styles.formGroupName}>{name}</h4>

                { subtitle && <span className="text-muted">{subtitle}</span> }
                { error && <span className={styles.formGroupError}>*{error}</span>}
            </div>

            <div>{children}</div>

            { description &&
                <span className={classnames(styles.formGroupDescription, 'text-muted')}>
                    {description}
                </span>
            }
        </fieldset>
    )
}
