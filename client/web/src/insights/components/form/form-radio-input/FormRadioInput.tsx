import classnames from 'classnames'
import React, { InputHTMLAttributes } from 'react'

import { RadioButton } from '@sourcegraph/wildcard'

interface RadioInputProps extends InputHTMLAttributes<HTMLInputElement> {
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
export const FormRadioInput: React.FunctionComponent<RadioInputProps> = ({
    title,
    description,
    className,
    labelTooltipText,
    labelTooltipPosition,
    name,
    value,
    checked,
    onChange,
    disabled,
}) => {
    const radioProps = { name, value, checked, onChange, disabled }
    return (
        <span
            data-placement={labelTooltipPosition}
            data-tooltip={labelTooltipText}
            className={classnames('d-flex flex-wrap align-items-center', className)}
        >
            <RadioButton
                aria-label={title}
                {...radioProps}
                labelProps={{
                    className: classnames({ 'text-muted': disabled }),
                }}
            />

            <span className="pl-2">{title}</span>

            {description && (
                <>
                    <span className="pl-2 pr-2">—</span>
                    <span className="text-muted">{description}</span>
                </>
            )}
        </span>
    )
}
