import React, { type InputHTMLAttributes } from 'react'

import classNames from 'classnames'

import { Label, Tooltip, type TooltipProps } from '@sourcegraph/wildcard'

interface RadioInputProps extends InputHTMLAttributes<HTMLInputElement> {
    /** Title of radio input. */
    title: string
    /** Description text for radio input. */
    description?: string
    /** Custom class name for root label element. */
    className?: string
    /** Tooltip text for radio label element. */
    labelTooltipText?: string
    /** Tooltip position */
    labelTooltipPosition?: TooltipProps['placement']
}

/** Displays form radio input for code insight creation form. */
export const FormRadioInput: React.FunctionComponent<React.PropsWithChildren<RadioInputProps>> = props => {
    const { title, description, className, labelTooltipText, labelTooltipPosition, ...otherProps } = props

    return (
        <Tooltip content={labelTooltipText} placement={labelTooltipPosition}>
            <Label
                className={classNames('d-flex flex-wrap align-items-center', className, {
                    'text-muted': otherProps.disabled,
                })}
            >
                {/* eslint-disable-next-line react/forbid-elements */}
                <input type="radio" {...otherProps} />

                <span className="pl-2">{title}</span>

                {description && (
                    <>
                        <span className="pl-2 pr-2">—</span>
                        <span className="text-muted">{description}</span>
                    </>
                )}
            </Label>
        </Tooltip>
    )
}
