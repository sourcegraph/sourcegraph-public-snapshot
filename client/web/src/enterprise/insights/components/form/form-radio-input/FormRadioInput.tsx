import classnames from 'classnames'
import React, { InputHTMLAttributes } from 'react'

import { RadioButton } from '@sourcegraph/wildcard'

interface RadioInputProps extends InputHTMLAttributes<HTMLInputElement> {
    /** Id of radio input. */
    id: string
    /** Name of radio input. */
    name: string
    /** Title of radio input. */
    title: string
    /** Description text for radio input. */
    description?: string
    /** Custom class name for root label element. */
    className?: string
    /** Tooltip text for radio label element. */
    labelTooltipText?: string
    /** Tooltip position */
    labelTooltipPosition?: string
}

/** Displays form radio input for code insight creation form. */
export const FormRadioInput: React.FunctionComponent<RadioInputProps> = props => {
    const radioProps = {
        value: props.value,
        checked: props.checked,
        onChange: props.onChange,
        disabled: props.disabled,
    }

    const label = (
        <>
            <span className="pl-2">{props.title}</span>
            {props.description && (
                <>
                    <span className="pl-2 pr-2">—</span>
                    <span className="text-muted">{props.description}</span>
                </>
            )}
        </>
    )
    return (
        <RadioButton
            data-placement={props.labelTooltipPosition}
            data-tooltip={props.labelTooltipText}
            id={props.id}
            inputProps={{ ...radioProps }}
            labelProps={{ className: classnames('', props.className, { 'text-muted': props.disabled }) }}
            label={label}
            name={props.name}
        />
    )
}
