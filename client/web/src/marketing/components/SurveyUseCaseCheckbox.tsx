import React, { useState } from 'react'

import classNames from 'classnames'

import { Button, ButtonProps } from '@sourcegraph/wildcard'

import styles from './SurveyUseCaseCheckbox.module.scss'

export interface SurveyUseCaseCheckboxProps extends Omit<ButtonProps, 'onChange'> {
    id: string
    label: React.ReactNode
    checked: boolean
    onChange: () => void
}

export const SurveyUseCaseCheckbox: React.FunctionComponent<SurveyUseCaseCheckboxProps> = ({
    id,
    label,
    onChange,
    checked,
    ...props
}) => {
    const [focused, setFocused] = useState<boolean>(false)

    return (
        <Button
            outline={!checked}
            variant={checked ? 'primary' : 'secondary'}
            size="sm"
            className={classNames(
                'd-flex align-items-center mb-0',
                styles.checkButton,
                checked && styles.checkButtonActive,
                focused && 'focus'
            )}
            as="label"
            {...props}
        >
            <span className={classNames(styles.checkbox, checked ? styles.checkmark : styles.checkboxDefault)} />
            {/* eslint-disable-next-line react/forbid-elements */}
            <input
                onBlur={() => setFocused(false)}
                onFocus={() => setFocused(true)}
                id={id}
                type="checkbox"
                checked={checked}
                onChange={onChange}
                className={styles.usecaseCheck}
            />
            <span className={classNames('ml-1', checked && styles.checkboxLabelActive)}>{label}</span>
        </Button>
    )
}
