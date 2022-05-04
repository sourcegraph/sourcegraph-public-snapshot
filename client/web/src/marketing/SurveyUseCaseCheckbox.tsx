import React, { useState } from 'react'

import classNames from 'classnames'

import { Button, ButtonProps, Checkbox } from '@sourcegraph/wildcard'

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
                {
                    focus: focused,
                }
            )}
            as="label"
            {...props}
        >
            <span className={classNames(styles.checkbox, checked ? styles.checkmark : styles.checkboxDefault)} />
            <Checkbox
                onBlur={() => setFocused(false)}
                onFocus={() => setFocused(true)}
                label={
                    <span
                        id={id}
                        className={classNames('ml-1', styles.checkboxLabel, checked && styles.checkboxLabelActive)}
                    >
                        {label}
                    </span>
                }
                id={id}
                checked={checked}
                onChange={onChange}
                wrapperClassName={styles.checkboxFormCheck}
                className={styles.usecaseCheck}
            />
        </Button>
    )
}
