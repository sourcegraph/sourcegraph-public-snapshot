import React, { FunctionComponent } from 'react'

import { defaultDurationValues } from './shared'

export interface DurationSelectProps {
    id: string
    value: string
    disabled: boolean
    onChange?: (value: number) => void
    durationValues?: { value: number; displayText: string }[]
}

export const DurationSelect: FunctionComponent<DurationSelectProps> = ({
    id,
    value,
    disabled,
    onChange,
    durationValues = defaultDurationValues,
}) => (
    <select
        id={id}
        className="form-control"
        value={value}
        disabled={disabled}
        onChange={event => onChange?.(Math.floor(parseInt(event.target.value, 10)))}
    >
        <option value="">Select duration</option>

        {durationValues.map(({ value, displayText }) => (
            <option key={value} value={value}>
                {displayText}
            </option>
        ))}
    </select>
)
