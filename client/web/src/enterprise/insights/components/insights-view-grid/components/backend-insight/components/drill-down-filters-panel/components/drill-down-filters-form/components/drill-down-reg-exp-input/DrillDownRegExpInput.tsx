import React, { ChangeEvent, forwardRef, InputHTMLAttributes, PropsWithChildren, Ref } from 'react'

import classNames from 'classnames'

import { Button, FlexTextArea, Label } from '@sourcegraph/wildcard'

import { TruncatedText } from '../../../../../../../../../trancated-text/TrancatedText'

import styles from './DrillDownRegExpInput.module.scss'

interface DrillDownRegExpInputProps extends InputHTMLAttributes<HTMLInputElement> {
    prefix: string
}

export const DrillDownRegExpInput = forwardRef((props: DrillDownRegExpInputProps, reference: Ref<HTMLInputElement>) => {
    const { prefix, ...inputProps } = props

    return (
        <span className="d-flex w-100">
            <span className={classNames(styles.prefixText, 'text-monospace')}>{prefix}</span>
            <FlexTextArea
                {...inputProps}
                containerClassName="w-100"
                className={classNames(inputProps.className, styles.input)}
                ref={reference}
            />
        </span>
    )
})

interface DrillDownContextInputProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'value' | 'onChange'> {
    value: string[]
    onChange: (value: string[]) => void
}

export const DrillDownContextInput = forwardRef<HTMLInputElement, DrillDownContextInputProps>((props, reference) => {
    const { value, onChange, ...attributes } = props

    const handleChange = (event: ChangeEvent<HTMLInputElement>): void => {
        const { value } = event.target

        onChange(value.split(',').filter(part => !!part))
    }

    return <DrillDownRegExpInput {...attributes} prefix="context:" value={value.join(',')} onChange={handleChange} />
})

DrillDownRegExpInput.displayName = 'DrillDownRegExpInput'

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
