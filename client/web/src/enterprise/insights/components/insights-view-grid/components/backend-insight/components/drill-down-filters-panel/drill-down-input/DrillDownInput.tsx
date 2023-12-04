import React, { forwardRef, type InputHTMLAttributes, type PropsWithChildren, type Ref } from 'react'

import classNames from 'classnames'

import { Button, Input, Label, type InputProps } from '@sourcegraph/wildcard'

import { TruncatedText } from '../../../../../../trancated-text/TruncatedText'

import styles from './DrillDownInput.module.scss'

interface DrillDownRegExpInputProps extends InputProps, InputHTMLAttributes<HTMLInputElement> {
    prefix: string
}

export const DrillDownInput = forwardRef((props: DrillDownRegExpInputProps, reference: Ref<HTMLInputElement>) => {
    const { prefix, className, ...inputProps } = props

    return (
        <span className={classNames(className, 'd-flex w-100')}>
            <span className={classNames(styles.prefixText, 'text-monospace')}>{prefix}</span>
            <Input
                {...inputProps}
                className={styles.inputContainer}
                inputClassName={classNames(styles.input)}
                ref={reference}
            />
        </span>
    )
})

DrillDownInput.displayName = 'DrillDownInput'

interface LabelWithResetProps {
    text: string
    disabled?: boolean
    className?: string
    onReset?: () => void
}

export const LabelWithReset: React.FunctionComponent<PropsWithChildren<LabelWithResetProps>> = props => (
    <Label className={classNames(styles.label, props.className)}>
        <span className={styles.labelText}>
            <TruncatedText>{props.text}</TruncatedText>

            <Button
                variant="link"
                size="sm"
                disabled={props.disabled}
                className={styles.labelResetButton}
                onClick={props.onReset}
            >
                Reset
            </Button>
        </span>

        {props.children}
    </Label>
)
