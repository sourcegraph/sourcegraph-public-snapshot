import classnames from 'classnames';
import React, { PropsWithChildren, ReactElement } from 'react';

import styles from './FormGroup.module.scss'

interface FormGroupProps {
    className?: string;
    name: string;
    subtitle?: string;
    description?: string;
}

export function FormGroup(props: PropsWithChildren<FormGroupProps>): ReactElement {
    const { className, name, subtitle, children, description } = props;

    return (
        <fieldset className={classnames(styles.formGroup, className, {[styles.formGroupWithSubtitle]: !!subtitle })}>

            <h4 className={styles.formGroupName}>{name}</h4>

            { subtitle && <p className="text-muted">{subtitle}</p> }

            <div className={styles.formGroupContent}>
                {children}
            </div>

            { description &&
                <span className={classnames(styles.formGroupDescription, 'text-muted')}>
                    {description}
                </span>
            }
        </fieldset>
    )
}
