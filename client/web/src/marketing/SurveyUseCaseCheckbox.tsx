import React, { useState } from 'react'

import classNames from 'classnames'

import { Button, ButtonProps } from '@sourcegraph/wildcard'

import styles from './SurveyUseCaseCheckbox.module.scss'

export interface SurveyUseCaseCheckboxProps extends Omit<ButtonProps, 'onChange'> {
    label: React.ReactNode
    onChange: (isChecked: boolean) => void
}

export const SurveyUseCaseCheckbox: React.FunctionComponent<SurveyUseCaseCheckboxProps> = ({
    label,
    onChange,
    ...props
}) => {
    const [checked, setChecked] = useState(false)

    return (
        <Button
            outline={true}
            variant="secondary"
            size="sm"
            onClick={() => {
                onChange(!checked)
                setChecked(current => !current)
            }}
            className={classNames('d-flex align-items-center', styles.checkButton, checked && styles.checkButtonActive)}
            {...props}
        >
            <span className={classNames(styles.checkbox, checked ? styles.checkmark : styles.checkboxDefault)} />
            {label}
        </Button>
    )
}
